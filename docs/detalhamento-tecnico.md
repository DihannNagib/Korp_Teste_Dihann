# Detalhamento Técnico

Este documento responde, item a item, aos pontos de detalhamento técnico
exigidos no enunciado do teste, explicando o **porquê** de cada decisão.
Passo a passo de execução, endpoints e roteiros de teste estão nos
READMEs de cada módulo (linkados ao longo deste documento).

---

## 1. Visão Geral e Arquitetura

Arquitetura de microsserviços com dois serviços em Go (Estoque e
Faturamento), cada um com seu próprio banco PostgreSQL, e um frontend em
Angular consumindo ambos.

```text
                                              ┌──────────────────┐
                                              │     Angular      │
                                              │      :4200       │
                                              └────────┬─────────┘
                                                       │
                                         ┌─────────────┴─────────────┐
                                         │                           │
                                 ┌───────▼───────┐           ┌───────▼───────┐
                                 │    Estoque    │           │  Faturamento  │
                                 │     :8080     │◄──────────┤     :8081     │
                                 └───────┬───────┘   HTTP    └───────┬───────┘
                                         │                           │
                                 ┌───────▼───────┐           ┌───────▼───────┐
                                 │  PostgreSQL   │           │  PostgreSQL   │
                                 │     :5433     │           │     :5434     │
                                 └───────────────┘           └───────────────┘
```

**Por que dois bancos separados:** cada microsserviço é dono exclusivo do
seu schema. Isso evita acoplamento via banco compartilhado (uma mudança de
schema no Estoque não pode quebrar o Faturamento) e reflete a
independência de deploy que a arquitetura de microsserviços exige.

Detalhes de execução de cada serviço:
[`backend/estoque/README.md`](../backend/estoque/README.md) |
[`backend/faturamento/README.md`](../backend/faturamento/README.md).

---

## 2. Frontend (Angular)

> Seção a ser preenchida conforme o frontend for implementado.

### 2.1 Ciclos de vida do Angular utilizados

*A preencher.*

### 2.2 Uso da biblioteca RxJS

*A preencher — indicar se houve uso (ex: `Observable` em chamadas HTTP,
`Subject` para comunicação entre componentes, operadores como
`switchMap`/`catchError`) e por quê.*

### 2.3 Outras bibliotecas utilizadas

*A preencher.*

### 2.4 Bibliotecas de componentes visuais

*A preencher (ex: Angular Material, PrimeNG, ou componentes próprios).*

---

## 3. Backend (Go)

### 3.1 Gerenciamento de dependências

Go Modules (`go.mod` / `go.sum`), um módulo independente por
microsserviço — cada serviço declara e versiona só as dependências que
realmente usa, sem um `go.mod` compartilhado que acoplasse os dois
serviços.

### 3.2 Frameworks utilizados

- **Gin** — roteamento HTTP e middlewares (logging, recovery).
- **GORM** — ORM sobre PostgreSQL: queries, transações e locking.

### 3.3 Tratamento de erros e exceções no backend

O tratamento de erros segue uma separação clara por camada:

**Domínio** — regras de negócio retornam *sentinel errors* Go
(`errors.New`), sem qualquer conhecimento de HTTP ou banco. Exemplo: a
regra de saldo não pode ficar negativo vive em `Produto.AjustarSaldo`,
que retorna `domain.ErrSaldoInsuficiente`. Isso permite testar a regra
isoladamente, sem banco nem HTTP, e garante que ela vale independente de
qual camada estiver chamando.

```go
func (p *Produto) AjustarSaldo(delta int) error {
	novoSaldo := p.Saldo + delta
	if novoSaldo < 0 {
		return ErrSaldoInsuficiente
	}
	p.Saldo = novoSaldo
	return nil
}
```

**Repository** — traduz erros nativos de infraestrutura (driver do
Postgres) para erros de domínio/aplicação. Em vez de fazer uma checagem
prévia de duplicidade (que teria uma condição de corrida em cenário
concorrente), o repository deixa a constraint `UNIQUE` do banco arbitrar
e traduz o erro nativo do driver (`pgconn.PgError`, código `23505`) para
`repository.ErrCodigoJaExiste`:

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
	return ErrCodigoJaExiste
}
```

**Validação de entrada (bind HTTP)** — o Gin usa `validator/v10` via tags
de struct (`binding:"required"`, `binding:"gte=0"`) para rejeitar payload
malformado antes de qualquer chamada ao service. O erro nativo do
validator é traduzido para uma lista de erros por campo, em português,
consumível diretamente pelo frontend (formato de resposta documentado em
[`backend/estoque/README.md`](../backend/estoque/README.md#formato-de-erro)).

**Handler** — ponto único de tradução erro → HTTP. A função
`responderErro` centraliza todo o mapeamento: primeiro verifica se é erro
de validação (`validator.ValidationErrors`), depois usa `errors.Is` para
identificar erros de domínio/repository/service específicos e retornar o
status correto:

| Origem do erro | Status HTTP |
|---|---|
| `validator.ValidationErrors` (campo obrigatório/inválido) | `400` |
| `repository.ErrProdutoNaoEncontrado` | `404` |
| `repository.ErrCodigoJaExiste` | `409` |
| `domain.ErrSaldoInsuficiente` | `422` |
| Erros de validação de negócio do service | `400` |
| Não mapeado / desconhecido | `400` (payload) |

Centralizar essa tradução em um único ponto evita duplicar `switch`/`if`
de mapeamento de erro em cada endpoint, e documenta o contrato de erros
da API em um único lugar do código.

**Panics inesperados** — capturados por `gin.CustomRecovery`, que loga o
panic e responde `500` com JSON, garantindo que o processo nunca derruba
por um erro não tratado (ex: nil pointer).

**Health check como sinal de falha de dependência** — `/health` faz
`Ping()` real no banco a cada chamada e retorna `503` se a conexão
falhar, em vez de simplesmente responder `200` fixo. Isso é o que
possibilita observar e testar o cenário obrigatório de falha de
microsserviço (ver seção 7).

**Faturamento — mapeamento equivalente**, com uma nuance própria: erros
vindos da chamada ao Estoque durante a impressão são diferenciados em
três categorias — `ErrSaldoInsuficienteEstoque` (`422`),
`ErrProdutoNaoEncontradoEstoque` (`404`) e `ErrFalhaComunicacaoEstoque`
(`503`, cobrindo indisponibilidade de rede/timeout e qualquer status
HTTP não mapeado explicitamente). Essa diferenciação já existia no
`client.EstoqueClient` desde a primeira versão do serviço; o `service`
apenas repassa a classificação original em vez de agrupar tudo sob um
único erro genérico. Detalhes:
[`backend/faturamento/README.md`](../backend/faturamento/README.md#mapeamento-de-erro--status-http).

### 3.4 Uso de LINQ / C#

Não aplicável — a solução foi implementada inteiramente em Go, sem uso de
C#.

---

## 4. Domínio de Estoque

A regra de saldo não-negativo vive no método `Produto.AjustarSaldo`, no
**domínio** — não no repository nem no handler. Essa escolha é o que
permite testar a regra de negócio isoladamente (ver
`internal/domain/produto_test.go`), sem depender de banco de dados ou
HTTP para validar o comportamento.

Estrutura completa da entidade `Produto`:
[`backend/estoque/README.md`](../backend/estoque/README.md#entidade-produto).

### 4.1 Persistência e Concorrência

O ajuste de saldo roda dentro de uma transação com bloqueio pessimista:

```sql
SELECT ... FROM produtos WHERE codigo = ? FOR UPDATE;
```

```text
Início da transação
        ↓
SELECT ... FOR UPDATE      (bloqueia a linha do produto)
        ↓
Produto.AjustarSaldo(delta)  (regra de domínio)
        ↓
UPDATE
        ↓
COMMIT (libera o lock)
```

Isso serializa o acesso ao mesmo produto: se duas requisições tentarem
baixar saldo do mesmo produto ao mesmo tempo, a segunda transação
aguarda a primeira commitar antes de ler o saldo — eliminando a
possibilidade de saldo negativo por condição de corrida.

Validado manualmente com duas requisições simultâneas contra um produto
de saldo `1`: uma responde `200`, a outra `422`, nunca as duas com
sucesso. Passo a passo:
[`backend/estoque/README.md`](../backend/estoque/README.md#teste-manual-de-concorrência).

## 4.2 Domínio de Faturamento

### 4.2.1 Entidades

`NotaFiscal` possui `numero` (numeração fiscal, exposta ao usuário) e
`id` (chave técnica) como campos deliberadamente separados, cada um
gerado por sua própria sequência no Postgres. Isso desacopla a
numeração sequencial exigida pelo enunciado de detalhes de
implementação da chave primária — se o schema precisar mudar no
futuro (ex: chave técnica virar UUID), a sequência de `numero`
continua intacta.

`ItemNotaFiscal` referencia a nota via chave estrangeira com
`ON DELETE CASCADE`, e tem `quantidade > 0` reforçado tanto no domínio
quanto por `CHECK` constraint no banco — mesma filosofia de dupla
camada de validação aplicada no Estoque.

### 4.2.2 Regras de negócio no domínio

Assim como em `Produto.AjustarSaldo` no Estoque, as regras de transição
de estado da nota vivem inteiramente no domínio, sem dependência de
banco ou HTTP:

```go
func (n *NotaFiscal) AdicionarItem(produtoCodigo string, quantidade int) error {
	if n.Status != StatusAberta {
		return ErrNotaNaoAberta
	}
	// ... validações de produtoCodigo e quantidade
}

func (n *NotaFiscal) PodeSerImpressa() error {
	if n.Status != StatusAberta {
		return ErrNotaNaoAberta
	}
	if len(n.Itens) == 0 {
		return ErrSemItens
	}
	return nil
}
```

Uma decisão específica do Faturamento: `Fechar()` retorna `error` e
recusa fechar uma nota que não esteja `ABERTA`, em vez de ser uma
operação incondicional. Isso obriga qualquer código chamador (o
`service`) a lidar explicitamente com esse retorno — reduzindo a
chance de a nota ser fechada duas vezes por engano em um fluxo com
múltiplas etapas.

### 4.2.3 Numeração sequencial

```sql
CREATE SEQUENCE IF NOT EXISTS notas_fiscais_numero_seq
    START WITH 1
    INCREMENT BY 1;

CREATE TABLE notas_fiscais (
    id BIGSERIAL PRIMARY KEY,
    numero BIGINT NOT NULL UNIQUE
        DEFAULT nextval('notas_fiscais_numero_seq'),
    ...
);
```

O número é atribuído atomicamente pelo Postgres no momento do insert,
eliminando qualquer possibilidade de duas notas concorrentes
receberem o mesmo número — o mesmo princípio de "deixar o banco
arbitrar" usado na constraint `UNIQUE` de código de produto no
Estoque (seção 3.3), aplicado agora a uma sequência em vez de uma
constraint de unicidade simples.

### 4.2.4 Persistência

`AtualizarStatus` faz um `UPDATE` direcionado apenas na coluna
`status`, em vez de `Save()` no struct completo — evita que o GORM
dispare upsert das associações (`Itens`) só porque elas estão
carregadas em memória no momento da chamada. Validado por teste de
integração dedicado (`TestNotaFiscalRepository_AtualizarStatusNaoAlteraItens`)
que confirma que os itens permanecem intactos após uma atualização de
status.

### 4.2.5 Integração com o Estoque e Compensação

A comunicação com o Estoque é feita via `EstoqueClient`
(`internal/client`), com dois métodos simétricos:

```go
func (c *EstoqueClient) BaixarItem(produtoCodigo string, quantidade int) error
func (c *EstoqueClient) EstornarItem(produtoCodigo string, quantidade int) error
```

Ambos usam o mesmo endpoint do Estoque (`PATCH /produtos/:codigo/saldo`),
diferindo apenas no sinal do `delta` — reaproveitando a decisão já
tomada no Estoque de ter um único endpoint de ajuste em vez de rotas
separadas de baixa/estorno.

O service depende da interface `EstoqueGateway`, não do client
concreto, permitindo testar `Imprimir` com um mock em testes unitários
sem subir o Estoque real.

**Baixa item a item com compensação (padrão saga):** ao imprimir,
`Imprimir` percorre os itens da nota individualmente, acumulando os
que já foram baixados com sucesso. Se um item falhar, todos os itens
já processados nesta tentativa são estornados antes do erro ser
propagado:

```go
processados := make([]domain.ItemNotaFiscal, 0, len(nota.Itens))

for _, item := range nota.Itens {
    if err := s.estoque.BaixarItem(item.ProdutoCodigo, item.Quantidade); err != nil {
        s.compensar(processados)
        return nil, mapearErroEstoque(err, item.ProdutoCodigo)
    }
    processados = append(processados, item)
}
```

Essa abordagem — baixar item a item e compensar em caso de falha, em
vez de delegar a baixa de todos os itens para dentro do client — foi
adotada porque só o `service` tem visibilidade de **quais** itens já
foram processados com sucesso no momento da falha; centralizar isso
no client exigiria que ele expusesse estado ou uma API mais complexa,
misturando responsabilidades de transporte HTTP com orquestração de
negócio.

**Garantia obtida:** ou a nota é processada por completo (todos os
itens baixados, nota `FECHADA`), ou nenhum saldo fica baixado
"órfão" — nunca um estado intermediário. Isso também torna a operação
de impressão **segura para retry**: uma nova tentativa após falha
reprocessa a nota do zero sem duplicar baixas de itens que já tinham
sido processados e depois compensados.

**Limitação aceita:** a compensação (`estornar`) é "melhor esforço" —
se o próprio estorno falhar (ex: o Estoque cai no meio da
compensação), o erro é registrado em log para reconciliação manual,
não retentado automaticamente nem persistido como pendência. Uma
evolução possível seria uma fila/tabela de compensações pendentes
reprocessadas por job assíncrono, fora do escopo desta versão.

Validado por teste unitário
(`TestImprimir_FalhaNoMeioCompensaItensAnteriores`,
`TestImprimir_RetryAposCompensacaoNaoDuplicaBaixa`) e por roteiro
manual com os dois serviços reais:
[`backend/faturamento/README.md`](../backend/faturamento/README.md#cenário-de-falha-requisito-obrigatório).

Detalhes de execução, schema completo e testes:
[`backend/faturamento/README.md`](../backend/faturamento/README.md).

---

## 5. Infraestrutura

### 5.1 Banco de dados

PostgreSQL 15, uma instância por microsserviço, cada uma em container
Docker independente com volume nomeado próprio (`estoque_data`,
`faturamento_data`), garantindo persistência entre reinícios dos
containers.

### 5.2 Migrations

Schema versionado via `golang-migrate`, com arquivos `.up.sql`/`.down.sql`
por microsserviço — permite aplicar/reverter mudanças de schema de forma
controlada e rastreável, em vez de depender de `AutoMigrate` em runtime.
Comandos: [`backend/estoque/README.md`](../backend/estoque/README.md#migrations).

### 5.3 Configuração

Variáveis de ambiente prefixadas por serviço (`ESTOQUE_*`,
`FATURAMENTO_*`), carregadas via `godotenv` em desenvolvimento local; em
produção, fornecidas diretamente pela infraestrutura de execução.

Os dois serviços divergem propositalmente na política de ausência de
variável: o Estoque aplica um valor padrão (`getEnv` com fallback), o
Faturamento falha imediatamente na inicialização (`requiredEnv` com
`log.Fatalf`). Essa segunda abordagem foi escolhida para o Faturamento
porque ele depende de uma variável crítica adicional
(`ESTOQUE_SERVICE_URL`) cuja ausência silenciosa levaria a um erro
difícil de diagnosticar só na hora da impressão da nota — preferimos que
o processo nem suba se a configuração estiver incompleta.

---

## 6. Requisitos Opcionais Implementados

- **Tratamento de concorrência:** lock pessimista (`SELECT ... FOR
  UPDATE`) descrito na seção 4.1, validado manualmente.
- **Idempotência:** implementada parcialmente via compensação
  automática (seção 4.2.5) — a operação de impressão é segura para
  retry após falha, pois qualquer baixa parcial é estornada antes do
  erro ser propagado. Não há, porém, uma chave de idempotência
  explícita (ex: header `Idempotency-Key`) nas chamadas HTTP entre os
  serviços; a garantia atual vem inteiramente do padrão de
  compensação no nível de aplicação, não de deduplicação na camada de
  transporte.
- **Uso de IA:** não implementado nesta versão.

---

## 7. Requisito Obrigatório: Tratamento de Falhas

### Falha de banco de dados (Estoque e Faturamento)

O cenário de falha de microsserviço é observável e testável através do
`/health` de cada serviço, que verifica a conexão real com o banco
(`sqlDB.Ping()`) a cada chamada, em vez de responder `200` fixo:

- Banco disponível → `200 {"status":"ok","database":"up"}`
- Banco indisponível → `503 {"status":"error","database":"down"}`

A aplicação se recupera sozinha assim que a conexão volta a responder,
sem necessidade de reiniciar o processo. Passo a passo de validação:
[`backend/estoque/README.md`](../backend/estoque/README.md#teste-de-falha-do-banco).

### Falha na comunicação Faturamento → Estoque

Durante a impressão de uma nota, se uma chamada ao Estoque falhar (por
indisponibilidade, saldo insuficiente ou produto não encontrado), o
Faturamento:

1. Estorna automaticamente qualquer item já baixado com sucesso nesta
   tentativa (compensação — ver seção 4.2.5);
2. Não fecha a nota — ela permanece `ABERTA`;
3. Devolve ao cliente HTTP um status diferenciado conforme a causa
   (`422` saldo insuficiente, `404` produto não encontrado, `503`
   indisponibilidade);
4. Permite nova tentativa de impressão assim que o problema for
   resolvido, sem risco de dupla baixa nos itens já processados
   anteriormente.

Validado manualmente com dois cenários: falha total (Estoque fora do
ar) e falha parcial (um item específico inválido no meio de uma nota
com múltiplos itens), confirmando em ambos os casos que a nota
permanece `ABERTA`, o saldo dos itens não é duplicado em retry, e o
saldo do item afetado pela falha volta ao valor original via
compensação. Roteiro completo:
[`backend/faturamento/README.md`](../backend/faturamento/README.md#cenário-de-falha-requisito-obrigatório).

---

## 8. Metodologia de Testes

- **Estoque** — automatizados (`go test ./... -v`): unitários de
  domínio (sem infraestrutura) e de integração de repository (contra
  PostgreSQL real). Manuais: roteiro end-to-end de 18 cenários +
  concorrência. Roteiro completo:
  [`backend/estoque/README.md`](../backend/estoque/README.md#roteiro-de-testes-manuais).
- **Faturamento** — automatizados: unitários de domínio, integração
  de repository, e unitários de service com mock de `EstoqueGateway`
  (incluindo compensação e retry após falha parcial). Manuais:
  roteiro de 21 cenários cobrindo CRUD de notas, transições de
  status, e os dois cenários obrigatórios de falha (total e parcial)
  do Estoque. Roteiro completo:
  [`backend/faturamento/README.md`](../backend/faturamento/README.md#roteiro-de-testes-manuais-completo).

---

## 9. Estado Atual da Implementação

### Concluído

- Arquitetura de microsserviços (Estoque + Faturamento) com bancos
  independentes;
- Serviço de Estoque completo: domínio, repository, service, handler,
  validação de payload, tratamento de erros centralizado, testes
  automatizados e roteiro de testes manuais (18 cenários + concorrência)
  validados;
- Serviço de Faturamento completo: domínio (`NotaFiscal` com
  numeração sequencial via sequência própria do Postgres), repository,
  cliente HTTP de integração com o Estoque, service com orquestração
  de impressão e **compensação automática** (padrão saga) em caso de
  falha parcial, handler com tratamento de erro diferenciado por
  status HTTP. Testes automatizados de domínio, repository e service
  (incluindo compensação e retry), e roteiro de 21 cenários manuais,
  cobrindo os dois cenários obrigatórios de falha (total e parcial)
  entre os dois serviços;
- Migrations versionadas de ambos os serviços;
- Health check de ambos os serviços com verificação real de banco,
  validado em cenário de falha e recuperação.

### Limitações conhecidas

- A compensação (estorno) em caso de falha parcial na impressão é
  "melhor esforço": se o próprio estorno falhar, o erro é apenas
  registrado em log, sem retry automático ou persistência de
  pendência para reconciliação (ver seção 4.2.5);
- Não há chave de idempotência explícita (`Idempotency-Key`) nas
  chamadas HTTP entre Faturamento e Estoque — a garantia de segurança
  em retry vem da compensação no nível de aplicação, não de um
  mecanismo de deduplicação na camada de transporte.

### Em andamento

- Dockerfile e integração do Faturamento no `docker-compose.yml`;
- Frontend Angular (telas de cadastro, listagem e impressão de notas).