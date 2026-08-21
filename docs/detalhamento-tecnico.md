# Detalhamento Técnico

Este documento apresenta as principais decisões técnicas da solução desenvolvida para o desafio de emissão de Notas Fiscais.

A aplicação é composta por um frontend em Angular e dois microsserviços em Go: **Estoque** e **Faturamento**.

Os detalhes de execução, endpoints e roteiros de testes estão disponíveis nos READMEs de cada módulo:

- [Backend — Estoque](../backend/estoque/README.md)
- [Backend — Faturamento](../backend/faturamento/README.md)
- [Frontend — Angular](../frontend/README.md)

---

## 1. Arquitetura

Foi utilizada uma arquitetura de microsserviços com separação de responsabilidades:

- **Estoque** — cadastro de produtos e controle de saldos.
- **Faturamento** — criação, consulta e impressão de notas fiscais.
- **Angular** — interface da aplicação e comunicação com os serviços.
- **PostgreSQL** — banco independente para cada microsserviço.

```text
                         ┌──────────────────┐
                         │     Angular      │
                         │      :4200       │
                         └────────┬─────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │                           │
             ┌──────▼───────┐           ┌──────▼────────┐
             │    Estoque   │   HTTP    │  Faturamento  │
             │     :8080    │◄──────────│     :8081     │
             └──────┬───────┘           └──────┬────────┘
                    │                           │
             ┌──────▼───────┐           ┌──────▼────────┐
             │  PostgreSQL  │           │  PostgreSQL   │
             │     :5433    │           │     :5434     │
             └──────────────┘           └───────────────┘
```

Cada serviço possui seu próprio banco, evitando compartilhamento de schema e reduzindo o acoplamento entre os microsserviços.

A comunicação entre Faturamento e Estoque é realizada via HTTP.

---

# 2. Frontend — Angular

## 2.1 Ciclos de vida

Foi utilizado principalmente o `ngOnInit`.

Ele é utilizado para inicialização dos componentes e carregamento dos dados necessários:

- `ProdutoListComponent` — carrega os produtos.
- `NotaListComponent` — carrega as notas.
- `NotaDetailComponent` — obtém o número da nota pela rota e carrega seus dados.
- `NotaFormComponent` — carrega os produtos disponíveis para inclusão na nota.

O `ngOnDestroy` não foi necessário, pois as chamadas HTTP utilizadas são observables de execução única que completam automaticamente.

---

## 2.2 RxJS

Foi utilizado RxJS através do `HttpClient`, que trabalha com `Observable`.

Principais recursos utilizados:

- `finalize()` — controla estados de carregamento tanto em sucesso quanto em erro.
- `tap()` — utilizado para efeitos colaterais, como emissão de eventos após a impressão.
- `catchError()` — utilizado no interceptor HTTP para logging centralizado.
- `Subject` / `Observable` — utilizados para comunicação de eventos entre componentes.

---

## 2.3 Outras bibliotecas e recursos

### Reactive Forms

Utilizado nos formulários de produtos e notas fiscais.

O formulário de nota utiliza `FormArray` para permitir múltiplos produtos e respectivas quantidades.

Também são utilizadas validações para campos obrigatórios, quantidade mínima, existência e duplicidade de produtos.

### Angular Router

Utilizado para navegação e para obtenção do número da nota através dos parâmetros da rota.

### Angular HttpClient

Utilizado para comunicação com os microsserviços.

---

## 2.4 Componentes visuais

Foi utilizado **Angular Material**.

Principais módulos:


| Módulo                     | Finalidade                   |
| -------------------------- | ---------------------------- |
| `MatToolbarModule`         | Navegação                    |
| `MatSidenavModule`         | Menu lateral                 |
| `MatListModule`            | Itens de navegação           |
| `MatTableModule`           | Listagens                    |
| `MatCardModule`            | Organização das informações  |
| `MatFormFieldModule`       | Campos de formulário         |
| `MatInputModule`           | Entrada de dados             |
| `MatAutocompleteModule`    | Sugestão de produtos         |
| `MatButtonModule`          | Ações                        |
| `MatIconModule`            | Ícones                       |
| `MatChipsModule`           | Status das notas             |
| `MatProgressSpinnerModule` | Indicadores de processamento |


O `MatProgressSpinner` é utilizado durante a impressão, atendendo ao requisito de exibir um indicador de processamento.

---

# 3. Backend — Go

## 3.1 Gerenciamento de dependências

Foi utilizado **Go Modules**, através dos arquivos `go.mod` e `go.sum`.

Cada microsserviço possui seu próprio módulo Go e gerencia suas dependências de forma independente.

---

## 3.2 Frameworks e bibliotecas

- **Gin** — framework HTTP, utilizado para rotas, handlers e middlewares.
- **GORM** — ORM utilizado para acesso ao PostgreSQL, transações e locking.
- **PostgreSQL** — persistência dos dados.
- **golang-migrate** — versionamento do schema através de migrations.
- **validator/v10** — validação dos payloads recebidos pela API.
- **godotenv** — carregamento das variáveis de ambiente em desenvolvimento local.

---

## 3.3 C# / LINQ

Não se aplica.

O backend foi desenvolvido integralmente em Go, sem utilização de C# ou LINQ.

---

# 4. Tratamento de erros

O tratamento de erros foi separado por camadas.

### Domínio

As regras de negócio retornam erros específicos (`sentinel errors`), sem dependência de HTTP ou banco de dados.

Exemplo: `ErrSaldoInsuficiente`.

### Repository

Erros de infraestrutura são traduzidos para erros conhecidos pela aplicação.

Por exemplo, a violação da constraint `UNIQUE` do PostgreSQL (`23505`) é convertida para `ErrCodigoJaExiste`.

A constraint do banco é utilizada como fonte de verdade, evitando uma verificação prévia de duplicidade sujeita a condição de corrida.

### Validação

Os payloads são validados antes da execução das regras de negócio através das validações do `validator/v10`.

### Handler

Os handlers traduzem os erros da aplicação para códigos HTTP.

Principais respostas do Estoque:


| Situação               | HTTP  |
| ---------------------- | ----- |
| Payload inválido       | `400` |
| Produto não encontrado | `404` |
| Código já existente    | `409` |
| Saldo insuficiente     | `422` |
| Erro inesperado        | `500` |


No Faturamento, erros provenientes do Estoque também são classificados como `404`, `422` ou `503`, conforme a causa.

### Panics

O Gin utiliza `CustomRecovery` para capturar panics inesperados, registrar o erro e retornar `500`.

---

# 5. Domínio de Estoque

O Estoque é responsável pelo cadastro dos produtos e controle dos saldos.

A regra de negócio de que o saldo não pode ficar negativo está no domínio, através de `Produto.AjustarSaldo`.

Isso permite testar a regra sem dependência de banco ou HTTP.

## Concorrência

Como requisito opcional, foi implementado **locking pessimista** com `SELECT ... FOR UPDATE` dentro de uma transação.

Isso serializa operações concorrentes sobre o mesmo produto.

Assim, considerando saldo `1` e duas operações simultâneas consumindo `1` unidade:

- uma operação reduz o saldo para `0`;
- a outra aguarda o lock;
- ao prosseguir, encontra saldo insuficiente;
- somente uma operação é concluída com sucesso.

---

# 6. Domínio de Faturamento

O Faturamento é responsável pelas notas fiscais e seus itens.

A nota possui:

- `ID` — identificador técnico;
- `Número` — numeração fiscal;
- `Status` — `ABERTA` ou `FECHADA`;
- `Itens` — produtos e quantidades.

A numeração é gerada pelo PostgreSQL através de uma `SEQUENCE`, garantindo geração sequencial sem depender de lógica de aplicação.

As principais regras de negócio são:

- somente notas `ABERTAS` podem receber itens;
- a nota deve possuir pelo menos um item para ser impressa;
- quantidade deve ser maior que zero;
- notas `FECHADAS` não podem ser impressas novamente;
- a nota só é fechada após o processamento de todos os itens.

---

# 7. Integração Faturamento → Estoque

Durante a impressão, o Faturamento utiliza um `EstoqueClient` para realizar as operações no microsserviço de Estoque.

O service depende da interface `EstoqueGateway`, permitindo utilizar mocks nos testes.

## Compensação

Os itens são processados individualmente.

Se uma operação posterior falhar, os itens que já foram baixados são estornados e a nota permanece `ABERTA`.



---

# 8. Banco de dados e Migrations

Foi utilizado **PostgreSQL 15**, com banco independente para cada microsserviço.


| Serviço     | Porta  |
| ----------- | ------ |
| Estoque     | `5433` |
| Faturamento | `5434` |


Os dados utilizam volumes Docker nomeados:

- `estoque_data`;
- `faturamento_data`.

O schema é versionado através do **golang-migrate**, utilizando migrations `.up.sql` e `.down.sql`.

A aplicação não depende de `AutoMigrate` durante a execução.

---

# 9. Configuração

As configurações são fornecidas através de variáveis de ambiente, separadas por serviço:

```text
ESTOQUE_*
FATURAMENTO_*
```

O `godotenv` é utilizado em desenvolvimento local.

Em ambientes de produção, as variáveis podem ser fornecidas diretamente pela infraestrutura.

Informações sensíveis não são armazenadas no código-fonte.

---

# 10. Tratamento de Falhas

O tratamento de falhas atende ao requisito obrigatório do desafio.

## 10.1 Falha do banco

Os dois microsserviços possuem `/health`, que realiza um `Ping()` real no banco.

Quando o banco está disponível, o endpoint retorna `200`.

Quando o banco está indisponível, retorna `503`.

A aplicação continua em execução e volta a responder normalmente quando a conexão com o banco é restabelecida.

## 10.2 Falha Faturamento → Estoque

Durante a impressão, se o Estoque estiver indisponível:

1. a comunicação falha;
2. itens já processados são compensados;
3. a nota permanece `ABERTA`;
4. o Faturamento retorna `503`;
5. o frontend apresenta uma mensagem apropriada ao usuário.

Para erros de negócio do Estoque:

- `404` — produto não encontrado;
- `422` — saldo insuficiente;
- `503` — indisponibilidade ou falha de comunicação.

---

# 11. Requisitos opcionais

## Concorrência

**Implementado.**

Foi utilizado `SELECT ... FOR UPDATE` no controle de saldo.

## Inteligência Artificial

**Não implementado.**

A funcionalidade era opcional e não foi adicionada sem uma necessidade de negócio relevante para o escopo.

## Idempotência

**Não implementada formalmente.**

Não foi utilizado `Idempotency-Key`.

Foi implementado mecanismo de compensação para evitar efeitos parciais durante a impressão.

A compensação reduz o risco de efeitos duplicados em uma nova tentativa, mas não substitui uma implementação formal de idempotência.

---

# 12. Testes

## Estoque

Foram implementados:

- testes unitários de domínio;
- testes de integração do repository com PostgreSQL real;
- testes da API;
- testes manuais dos principais fluxos;
- teste de concorrência.

## Faturamento

Foram implementados:

- testes unitários de domínio;
- testes de integração do repository;
- testes do service com mock do `EstoqueGateway`;
- testes de compensação;
- testes de retry após falha parcial;
- testes da API;
- testes manuais dos principais fluxos;
- cenários de falha total e parcial do Estoque.

## Frontend

O frontend foi validado através de testes manuais, cobrindo:

- cadastro e listagem de produtos;
- criação de notas;
- múltiplos itens;
- consulta;
- impressão;
- indicador de processamento;
- fechamento;
- bloqueio de reimpressão;
- saldo insuficiente;
- produto inexistente;
- indisponibilidade do Estoque.

Os roteiros completos estão nos READMEs de cada módulo.

---

# 13. Resumo


| Requisito                  | Implementação                                            |
| -------------------------- | -------------------------------------------------------- |
| Frontend                   | Angular                                                  |
| Lifecycle                  | `ngOnInit`                                               |
| Formulários                | Reactive Forms                                           |
| Comunicação HTTP           | Angular `HttpClient`                                     |
| RxJS                       | `Observable`, `Subject`, `finalize`, `tap`, `catchError` |
| Componentes visuais        | Angular Material                                         |
| Backend                    | Go                                                       |
| Dependências               | Go Modules                                               |
| Framework HTTP             | Gin                                                      |
| ORM                        | GORM                                                     |
| Banco                      | PostgreSQL 15                                            |
| Migrations                 | golang-migrate                                           |
| Validação                  | validator/v10                                            |
| Microsserviços             | Estoque + Faturamento                                    |
| Comunicação entre serviços | HTTP                                                     |
| Concorrência               | `SELECT ... FOR UPDATE`                                  |
| Numeração                  | PostgreSQL `SEQUENCE`                                    |
| Falhas                     | `/health` + compensação                                  |
| IA                         | Não implementada                                         |
| Idempotência formal        | Não implementada                                         |
| Testes backend             | Unitários + integração + API + manuais                   |
| Testes frontend            | Manuais                                                  |
| C# / LINQ                  | Não aplicável                                            |


---

# 14. Conclusão

A solução atende aos requisitos funcionais e arquiteturais do desafio, utilizando Angular, dois microsserviços independentes em Go e bancos PostgreSQL separados.

Além dos requisitos obrigatórios, foram implementados tratamento de concorrência e compensação de operações parcialmente concluídas durante a impressão.

As principais decisões foram orientadas por separação de responsabilidades, integridade dos dados, isolamento entre microsserviços, tratamento explícito de erros e testabilidade.

Os detalhes de execução, endpoints e roteiros completos de testes estão disponíveis nos READMEs específicos de cada módulo.