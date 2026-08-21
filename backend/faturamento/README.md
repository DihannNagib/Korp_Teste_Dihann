# Serviço de Faturamento

Microsserviço responsável pela emissão e gestão de notas fiscais.
Depende do serviço de Estoque para baixar saldo de produtos no momento
da impressão da nota, via chamada HTTP síncrona.

## Stack

- Go 1.21+
- [Gin](https://github.com/gin-gonic/gin) — roteamento HTTP
- [GORM](https://gorm.io) — acesso a dados / PostgreSQL
- [golang-migrate](https://github.com/golang-migrate/migrate) — versionamento de schema
- [testify](https://github.com/stretchr/testify) — testes automatizados

## Variáveis de ambiente

Lidas a partir do `.env` na raiz do repositório (prefixo `FATURAMENTO_`).

```env
FATURAMENTO_DB_HOST
FATURAMENTO_DB_PORT
FATURAMENTO_DB_USER
FATURAMENTO_DB_PASSWORD
FATURAMENTO_DB_NAME
FATURAMENTO_API_PORT
ESTOQUE_SERVICE_URL
```

## Subindo o banco

```bash
docker compose up -d db_faturamento
```

## Migrations

```bash
cd backend/faturamento

migrate -path migrations \
  -database "postgres://<FATURAMENTO_DB_USER>:<FATURAMENTO_DB_PASSWORD>@<FATURAMENTO_DB_HOST>:<FATURAMENTO_DB_PORT>/<FATURAMENTO_DB_NAME>?sslmode=disable" \
  up
```

Para reverter: `migrate -path migrations -database "..." down 1`.

### Schema

- **`notas_fiscais`** — `numero` preenchido via
  `DEFAULT nextval('notas_fiscais_numero_seq')`, sequência própria e
  independente do `id`, garantindo numeração sequencial atômica no
  banco.
- **`itens_nota_fiscal`** — `FOREIGN KEY ... ON DELETE CASCADE`;
  `CHECK (quantidade > 0)` reforça no banco a mesma regra do domínio.

## Executando o serviço

Com o Estoque já rodando em `:8080`:

```bash
cd backend/faturamento
go run ./cmd
```

- **API:** http://localhost:8081
- **Health check:** http://localhost:8081/health — `200` se o serviço e
  o banco estiverem disponíveis, `503` se a conexão com o banco falhar.

## Testes automatizados

```bash
cd backend/faturamento
go test ./... -v
```

- `internal/domain` — unitários puros: transições de estado da nota
  (`AdicionarItem`, `PodeSerImpressa`, `Fechar`) e suas validações.
- `internal/repository` — integração contra o PostgreSQL do `.env`
  (banco precisa estar rodando). Limpam os próprios dados via
  `t.Cleanup`.
- `internal/service` — unitários com mock de `EstoqueGateway` e do
  repository, incluindo o cenário de falha parcial na impressão (ver
  [Cenário de falha](#cenário-de-falha-requisito-obrigatório) abaixo).

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Status do serviço e da conexão com o banco |
| POST | `/api/v1/notas` | Cria nota fiscal (status inicial `ABERTA`) |
| GET | `/api/v1/notas` | Lista notas fiscais |
| GET | `/api/v1/notas/:numero` | Busca nota por número |
| POST | `/api/v1/notas/:numero/imprimir` | Imprime a nota: baixa saldo no Estoque e fecha a nota |

### Entidades

**`NotaFiscal`**

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | `uint` | Chave técnica |
| `numero` | `uint` | Numeração fiscal sequencial, gerada pelo banco |
| `status` | `string` | `ABERTA` ou `FECHADA` |
| `itens` | `[]ItemNotaFiscal` | Produtos e quantidades |
| `createdAt` / `updatedAt` | `time.Time` | Auditoria |

**`ItemNotaFiscal`**

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | `uint` | Identificador |
| `notaFiscalId` | `uint` | Referência à nota (FK) |
| `produtoCodigo` | `string` | Código do produto no Estoque |
| `quantidade` | `int` | Quantidade utilizada (`> 0`) |

### Exemplo — criar nota

```bash
curl -X POST http://localhost:8081/api/v1/notas \
  -H "Content-Type: application/json" \
  -d '{"itens":[{"produtoCodigo":"P001","quantidade":2},{"produtoCodigo":"P002","quantidade":1}]}'
```

### Exemplo — imprimir nota

```bash
curl -X POST http://localhost:8081/api/v1/notas/1/imprimir
```

### Regras de domínio implementadas

- Toda nota nasce `ABERTA` (`NovaNotaFiscal`).
- Itens só podem ser adicionados a notas `ABERTA`.
- Código de produto vazio ou quantidade `<= 0` são rejeitados no
  domínio, antes de qualquer persistência.
- Impressão exige nota `ABERTA` **e** ao menos um item
  (`PodeSerImpressa`).
- `Fechar()` recusa fechar nota que não esteja `ABERTA` (retorna
  `ErrNotaNaoAberta`), em vez de ser uma operação silenciosamente
  idempotente — obriga o service a checar o estado explicitamente.

### Formato de erro

Validação de payload:
```json
{"erros": [{"campo": "itens", "erro": "itens deve conter ao menos 1 item(ns)"}]}
```

Erro de negócio da nota:
```json
{"erro": "nota fiscal nao esta aberta"}
```

Erro vindo do Estoque durante a impressão (ver tabela abaixo para o
status correspondente):
```json
{"erro": "saldo insuficiente no estoque (produto P001)"}
```

### Mapeamento de erro → status HTTP

| Origem do erro | Status HTTP |
|---|---|
| `validator.ValidationErrors` | `400` |
| Número de nota inválido (não numérico ou zero) | `400` |
| `repository.ErrNotaNaoEncontrada` | `404` |
| `domain.ErrNotaNaoAberta` (imprimir nota já fechada) | `422` |
| `domain.ErrSemItens`, `ErrQuantidadeInvalida`, `ErrProdutoInvalido` | `400` |
| `service.ErrSaldoInsuficienteEstoque` | `422` |
| `service.ErrProdutoNaoEncontradoEstoque` | `404` |
| `service.ErrFalhaComunicacaoEstoque` (indisponibilidade/timeout) | `503` |
| Não mapeado | `500` |

O client de Estoque (`internal/client`) já distinguia esses três tipos
de erro internamente; o `service` repassa essa classificação ao handler
em vez de agrupar tudo sob um único erro genérico.

---

## Cenário de falha (requisito obrigatório)

Ao imprimir uma nota, o service baixa o saldo de cada item **um a um**
no Estoque. Se algum item falhar no meio do processo:

1. Todos os itens **já baixados com sucesso nesta tentativa** são
   automaticamente **estornados** (compensados) antes do erro ser
   propagado.
2. `Fechar()` / `AtualizarStatus` não são executados — a nota
   permanece `ABERTA`.
3. O erro é reportado ao usuário com o status HTTP correspondente
   (ver tabela acima).
4. Uma nova tentativa de impressão reprocessa a nota do zero, sem
   risco de dupla baixa nos itens que já tinham sido processados na
   tentativa anterior — porque eles foram estornados no passo 1.

Essa compensação (padrão *saga* / transação compensatória) garante que
o sistema nunca fica em um estado intermediário inconsistente: ou a
nota foi 100% processada (todos os itens baixados e nota `FECHADA`),
ou nenhum saldo permanece baixado e a nota continua `ABERTA` para nova
tentativa.

```go
for _, item := range nota.Itens {
    if err := s.estoque.BaixarItem(item.ProdutoCodigo, item.Quantidade); err != nil {
        s.compensar(processados) // estorna tudo que já foi baixado
        return nil, mapearErroEstoque(err, item.ProdutoCodigo)
    }
    processados = append(processados, item)
}
```

A compensação é "melhor esforço": se o próprio estorno falhar, a falha
é registrada em log para reconciliação manual, em vez de ser perdida
silenciosamente.

### Roteiro de teste manual — falha total do Estoque

```bash
curl -s -X POST localhost:8080/api/v1/produtos -H "Content-Type: application/json" \
  -d '{"codigo":"FALHA-001","descricao":"Produto A","saldo":10}'
curl -s -X POST localhost:8080/api/v1/produtos -H "Content-Type: application/json" \
  -d '{"codigo":"FALHA-002","descricao":"Produto B","saldo":10}'

curl -s -X POST localhost:8081/api/v1/notas -H "Content-Type: application/json" \
  -d '{"itens":[{"produtoCodigo":"FALHA-001","quantidade":3},{"produtoCodigo":"FALHA-002","quantidade":3}]}'
# anote o "numero"

# Derrube o Estoque (Ctrl+C)

curl -s -X POST localhost:8081/api/v1/notas/NUMERO/imprimir
# esperado: 503

curl -s localhost:8081/api/v1/notas/NUMERO
# esperado: status "ABERTA"

# Suba o Estoque novamente

curl -s -X POST localhost:8081/api/v1/notas/NUMERO/imprimir
# esperado: 200, "FECHADA"

curl -s localhost:8080/api/v1/produtos/FALHA-001
curl -s localhost:8080/api/v1/produtos/FALHA-002
# esperado: saldo 7 em ambos (10 - 3), nunca 4 -- confirma que não houve dupla baixa
```

### Roteiro de teste manual — falha parcial (item específico)

```bash
curl -s -X POST localhost:8080/api/v1/produtos -H "Content-Type: application/json" \
  -d '{"codigo":"COMP-001","descricao":"Produto A","saldo":10}'

curl -s -X POST localhost:8081/api/v1/notas -H "Content-Type: application/json" \
  -d '{"itens":[{"produtoCodigo":"COMP-001","quantidade":3},{"produtoCodigo":"COMP-002","quantidade":1}]}'
# anote o "numero" -- COMP-002 não existe no Estoque de propósito

curl -s -X POST localhost:8081/api/v1/notas/NUMERO/imprimir
# esperado: 404 {"erro":"produto nao encontrado no estoque (produto COMP-002)"}

curl -s localhost:8080/api/v1/produtos/COMP-001
# esperado: saldo 10 -- a baixa do primeiro item foi estornada automaticamente

curl -s -X POST localhost:8080/api/v1/produtos -H "Content-Type: application/json" \
  -d '{"codigo":"COMP-002","descricao":"Produto B","saldo":10}'

curl -s -X POST localhost:8081/api/v1/notas/NUMERO/imprimir
# esperado: 200, "FECHADA"

curl -s localhost:8080/api/v1/produtos/COMP-001
# esperado: saldo 7 (10 - 3, uma única vez)
```

---

## Roteiro de testes manuais (completo)

| # | Cenário | Esperado |
|---|---|---|
| 1 | `GET /health` | `200` |
| 2 | Banco do Faturamento conectado | health reflete `database: up` |
| 3 | Criar nota válida | `201` |
| 4 | Criar nota sem itens | `400` |
| 5 | Criar nota com quantidade `0` | `400` |
| 6 | Criar nota com `produtoCodigo` vazio | `400` |
| 7 | Listar notas | `200` |
| 8 | Buscar nota existente | `200` |
| 9 | Buscar nota inexistente | `404` |
| 10 | Buscar nota com número inválido (não numérico ou `0`) | `400` |
| 11 | Imprimir nota `ABERTA` com Estoque disponível | `200`, status `FECHADA` |
| 12 | Saldo é efetivamente reduzido no Estoque | conferido via `GET /produtos/:codigo` do Estoque |
| 13 | Nota muda `ABERTA → FECHADA` após impressão | `200` |
| 14 | Tentar imprimir nota já `FECHADA` | `422` |
| 15 | Saldo insuficiente em um item | `422`, item anterior compensado |
| 16 | Nota permanece `ABERTA` quando Estoque rejeita a baixa | confirmado via `GET` |
| 17 | Produto inexistente no Estoque | `404`, item anterior compensado |
| 18 | Estoque totalmente indisponível | `503` |
| 19 | Nota permanece `ABERTA` com Estoque fora do ar | confirmado via `GET` |
| 20 | Após Estoque voltar, impressão pode ser repetida sem dupla baixa | `200`, saldo correto |
| 21 | Múltiplos itens são baixados corretamente (fluxo sem falha) | saldo de cada item reduzido corretamente |

Detalhamento das decisões técnicas:
[`docs/detalhamento-tecnico.md`](../../docs/detalhamento-tecnico.md).