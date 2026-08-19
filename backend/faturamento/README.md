# erviço de Faturamento

Microsserviço responsável pela emissão e gestão de notas fiscais. Depende
do serviço de Estoque para consultar e baixar saldo de produtos no
momento da impressão (integração ainda não implementada — ver
[Estado atual](#estado-atual)).

## Stack

- Go 1.21+
- [Gin](https://github.com/gin-gonic/gin) — roteamento HTTP
- [GORM](https://gorm.io) — acesso a dados / PostgreSQL
- [golang-migrate](https://github.com/golang-migrate/migrate) — versionamento de schema
- [testify](https://github.com/stretchr/testify) — testes automatizados

## Variáveis de ambiente

Lidas a partir do `.env` na raiz do repositório (prefixo `FATURAMENTO_`).

Diferente do serviço de Estoque, aqui **todas as variáveis são
obrigatórias** — `config.Load` usa `requiredEnv`, que encerra o processo
(`log.Fatalf`) caso qualquer uma esteja ausente, em vez de aplicar um
valor padrão silencioso. A decisão é intencional: preferimos falhar
imediatamente na inicialização a subir o serviço com uma configuração
incompleta (ex: apontando sem querer para o banco errado por falta de
uma variável).

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

A partir da raiz do repositório:

```bash
docker compose up -d db_faturamento
```

## Migrations

Schema versionado em `migrations/*.up.sql` / `*.down.sql`.

```bash
cd backend/faturamento

migrate -path migrations \
  -database "postgres://<FATURAMENTO_DB_USER>:<FATURAMENTO_DB_PASSWORD>@<FATURAMENTO_DB_HOST>:<FATURAMENTO_DB_PORT>/<FATURAMENTO_DB_NAME>?sslmode=disable" \
  up
```

Para reverter a última migration:

```bash
migrate -path migrations -database "..." down 1
```

### Schema

- `notas_fiscais` — `numero` é preenchido pelo Postgres via
`DEFAULT nextval('notas_fiscais_numero_seq')`, usando uma sequência
própria e independente do `id` (chave técnica). Isso garante a
numeração sequencial exigida pelo enunciado de forma atômica no
banco, sem depender de lógica de aplicação sujeita a condição de
corrida — o mesmo princípio usado no Estoque para garantir
unicidade de código de produto via constraint em vez de checagem
prévia em código.
- `itens_nota_fiscal` — referencia `notas_fiscais` via `FOREIGN KEY ... ON DELETE CASCADE`; `CHECK (quantidade > 0)` reforça no banco a mesma regra que o domínio já aplica em `AdicionarItem`.

## Testes automatizados

```bash
cd backend/faturamento
go test ./... -v
```

- `internal/domain` — testes unitários puros, sem banco: cobrem as
transições de estado da nota (`AdicionarItem`, `PodeSerImpressa`,
`Fechar`) e suas validações.
- `internal/repository` — testes de integração contra o PostgreSQL
configurado no `.env` (precisa do banco rodando). Limpam os próprios
dados via `t.Cleanup` ao final de cada teste.

## Estado atual


| Camada                                          | Status       |
| ----------------------------------------------- | ------------ |
| `config` / `database`                           | Concluído    |
| `domain` (`NotaFiscal`, `ItemNotaFiscal`)       | Concluído    |
| Migrations                                      | Concluído    |
| `repository`                                    | Concluído    |
| `client` (integração HTTP com Estoque)          | Não iniciado |
| `service` (orquestração + compensação em falha) | Não iniciado |
| `handler` / `main.go` (API HTTP)                | Não iniciado |


O serviço ainda **não expõe API HTTP** — não há `/health` nem
endpoints de notas fiscais disponíveis até a conclusão das camadas de
`client`, `service` e `handler`. Este README será atualizado com
endpoints, exemplos de `curl` e roteiro de testes manuais assim que a
API estiver funcional.

### Entidades

`NotaFiscal`


| Campo                     | Tipo               | Descrição                                      |
| ------------------------- | ------------------ | ---------------------------------------------- |
| `id`                      | `uint`             | Identificador técnico (chave primária)         |
| `numero`                  | `uint`             | Numeração fiscal sequencial, gerada pelo banco |
| `status`                  | `string`           | `ABERTA` ou `FECHADA`                          |
| `itens`                   | `[]ItemNotaFiscal` | Produtos e quantidades da nota                 |
| `createdAt` / `updatedAt` | `time.Time`        | Auditoria                                      |


`ItemNotaFiscal`


| Campo           | Tipo     | Descrição                    |
| --------------- | -------- | ---------------------------- |
| `id`            | `uint`   | Identificador                |
| `notaFiscalId`  | `uint`   | Referência à nota (FK)       |
| `produtoCodigo` | `string` | Código do produto no Estoque |
| `quantidade`    | `int`    | Quantidade utilizada (`> 0`) |


### Regras de domínio implementadas

- Toda nota nasce com status `ABERTA` (`NovaNotaFiscal`).
- Itens só podem ser adicionados a notas `ABERTA` (`AdicionarItem`
retorna `ErrNotaNaoAberta` caso contrário).
- Código de produto vazio ou quantidade `<= 0` são rejeitados no
domínio, antes de qualquer persistência.
- Impressão exige nota `ABERTA` **e** ao menos um item
(`PodeSerImpressa`).
- `Fechar()` só é permitido em nota `ABERTA`; chamar em nota já
fechada retorna `ErrNotaNaoAberta` em vez de ser uma operação
silenciosamente idempotente — decisão deliberada para que o service
nunca "esqueça" de checar o estado antes de fechar.

Detalhamento das decisões técnicas (por que a regra vive no domínio, por
que a numeração usa sequência própria etc.):
`[docs/detalhamento-tecnico.md](../../docs/detalhamento-tecnico.md)`.