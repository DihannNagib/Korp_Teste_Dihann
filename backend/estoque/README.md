# Serviço de Estoque

Microsserviço responsável pelo cadastro de produtos e controle de saldo
em estoque. Não depende de nenhum outro serviço — é consumido pelo
Faturamento e pelo frontend.

## Stack

- Go 1.21+
- [Gin](https://github.com/gin-gonic/gin) — roteamento HTTP
- [GORM](https://gorm.io) — acesso a dados / PostgreSQL
- [golang-migrate](https://github.com/golang-migrate/migrate) — versionamento de schema
- [testify](https://github.com/stretchr/testify) — testes automatizados

## Variáveis de ambiente

Lidas a partir do `.env` na raiz do repositório (prefixo `ESTOQUE_`):

```env
ESTOQUE_DB_HOST
ESTOQUE_DB_PORT
ESTOQUE_DB_USER
ESTOQUE_DB_PASSWORD
ESTOQUE_DB_NAME
ESTOQUE_API_PORT
```

## Subindo o banco

A partir da raiz do repositório:

```bash
docker compose up -d db_estoque
```

## Migrations

Schema versionado em `migrations/*.up.sql` / `*.down.sql`.

```bash
cd backend/estoque

migrate -path migrations \
  -database "postgres://<ESTOQUE_DB_USER>:<ESTOQUE_DB_PASSWORD>@<ESTOQUE_DB_HOST>:<ESTOQUE_DB_PORT>/<ESTOQUE_DB_NAME>?sslmode=disable" \
  up
```

Para reverter a última migration:

```bash
migrate -path migrations -database "..." down 1
```

## Executando o serviço

```bash
cd backend/estoque
go run ./cmd
```

- **API:** http://localhost:8080
- **Health check:** http://localhost:8080/health — retorna `200` se o
  serviço e o banco estiverem disponíveis, `503` se a conexão com o
  banco falhar (ver [teste de falha do banco](#teste-de-falha-do-banco)).

## Testes automatizados

```bash
cd backend/estoque
go test ./... -v
```

Os testes de `internal/domain` são unitários puros (sem banco). Os testes
de `internal/repository` sobem contra o PostgreSQL configurado no `.env`
(precisa do banco rodando) e limpam os próprios dados de teste
(prefixo `TESTE-`) ao final de cada execução.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Status do serviço e da conexão com o banco |
| POST | `/api/v1/produtos` | Cria produto |
| GET | `/api/v1/produtos` | Lista produtos |
| GET | `/api/v1/produtos/:codigo` | Busca produto por código |
| PATCH | `/api/v1/produtos/:codigo/saldo` | Ajusta saldo (`delta` positivo ou negativo) |

### Entidade `Produto`

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | `uint` | Identificador |
| `codigo` | `string` | Código único do produto |
| `descricao` | `string` | Descrição |
| `saldo` | `int` | Saldo atual |
| `createdAt` / `updatedAt` | `time.Time` | Auditoria |

### Exemplo — criar produto

```bash
curl -X POST http://localhost:8080/api/v1/produtos \
  -H "Content-Type: application/json" \
  -d '{"codigo":"P001","descricao":"Caneta Azul","saldo":10}'
```

### Exemplo — ajustar saldo

```bash
curl -X PATCH http://localhost:8080/api/v1/produtos/P001/saldo \
  -H "Content-Type: application/json" \
  -d '{"delta":-2}'
```

### Formato de erro

Erros de validação de campo (payload incompleto/inválido):

```json
{"erros": [{"campo": "codigo", "erro": "codigo e obrigatorio"}]}
```

Erros de negócio (produto não encontrado, código duplicado, saldo
insuficiente):

```json
{"erro": "saldo insuficiente"}
```

---

## Roteiro de testes manuais

Além dos testes automatizados, o serviço foi validado manualmente via
`curl`/Postman contra a API rodando com o banco real. Este roteiro cobre
os fluxos que os testes automatizados não exercitam fim-a-fim (handler +
service + repository + banco juntos, incluindo persistência entre
reinícios do processo).

| # | Cenário | Requisição | Esperado |
|---|---|---|---|
| 1 | Health check | `GET /health` | `200` `{"status":"ok","database":"up"}` |
| 2 | Listar vazio/existente | `GET /api/v1/produtos` | `200` array |
| 3 | Criar produto | `POST /api/v1/produtos` `{"codigo":"PROD-001","descricao":"Produto teste","saldo":10}` | `201` produto criado |
| 4 | Persistência após restart | reiniciar o processo (`Ctrl+C` → `go run ./cmd`), depois `GET /api/v1/produtos/PROD-001` | produto continua com saldo `10` — comprova persistência real, não em memória |
| 5 | Saldo inicial negativo | `POST` com `"saldo": -10` | `400` |
| 6 | Código vazio | `POST` com `"codigo": ""` | `400` |
| 7 | Descrição vazia | `POST` com `"descricao": ""` | `400` |
| 8 | Código duplicado | repetir `POST` com `"codigo":"PROD-001"` | `409` `{"erro":"codigo de produto ja cadastrado"}` — comprova que a unicidade é garantida pela constraint do Postgres |
| 9 | Buscar existente | `GET /api/v1/produtos/PROD-001` | `200` |
| 10 | Buscar inexistente | `GET /api/v1/produtos/NAO-EXISTE` | `404` `{"erro":"produto nao encontrado"}` |
| 11 | Ajuste — entrada | `PATCH /api/v1/produtos/PROD-001/saldo` `{"delta":5}` | `200`, saldo `10 → 15` |
| 12 | Ajuste — saída | `{"delta":-3}` | `200`, saldo `15 → 12` |
| 13 | Saldo insuficiente | `{"delta":-100}` | `422` `{"erro":"saldo insuficiente"}` **e** `GET` do produto continua mostrando saldo `12` — comprova rollback da transação |
| 14 | Delta zero | `{"delta":0}` | `400` |
| 15 | Ajuste em produto inexistente | `PATCH /api/v1/produtos/NAO-EXISTE/saldo` | `404` |
| 16 | JSON malformado | `POST` com corpo inválido (`{"codigo":}`) | `400` |
| 17 | Persistência (2ª rodada) | reiniciar o processo novamente, `GET /api/v1/produtos/PROD-001` | produto continua existindo |
| 18 | Banco indisponível | ver [teste de falha do banco](#teste-de-falha-do-banco) | `503` |

### Teste de falha do banco

Comprova o requisito obrigatório de tratamento de falha de microsserviço,
com a API em execução:

```bash
# 1. Derrubar o banco
docker compose stop db_estoque

# 2. Confirmar que o health check reporta a falha
curl -s http://localhost:8080/health
# esperado: HTTP 503 {"status":"error","database":"down"}

# 3. Restaurar o banco
docker compose start db_estoque

# 4. Confirmar recuperação automática (sem reiniciar a API)
curl -s http://localhost:8080/health
# esperado: HTTP 200 {"status":"ok","database":"up"}
```

### Teste manual de concorrência

Validado localmente (não versionado como script, pois é uma ferramenta
de teste e não faz parte da aplicação):

1. Criar um produto com `saldo: 10`.
2. Disparar duas requisições `PATCH .../saldo` com `{"delta":-7}` **em
   paralelo**, contra o mesmo `codigo`.
3. Resultado observado: uma requisição responde `200` (saldo vai para
   `3`) e a outra responde `422` (`saldo insuficiente`) — nunca as duas
   com sucesso, e o saldo final nunca fica negativo.

Explicação de por que isso funciona (bloqueio pessimista):
[`docs/detalhamento-tecnico.md`](../../docs/detalhamento-tecnico.md#41-persistência-e-concorrência).