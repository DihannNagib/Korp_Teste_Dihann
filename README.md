# Korp — Teste Técnico

Aplicação web composta por um frontend em Angular e dois microsserviços backend desenvolvidos em Go:

- **Estoque** — responsável pelo controle de produtos e saldos.
- **Faturamento** — responsável pela gestão das notas fiscais.

Cada microsserviço possui seu próprio banco de dados PostgreSQL.

## Arquitetura

A solução utiliza uma arquitetura baseada em microsserviços:

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
                                 │     :8080     │           │     :8081     │
                                 └───────┬───────┘           └───────┬───────┘
                                         │                           │
                                 ┌───────▼───────┐           ┌───────▼───────┐
                                 │  PostgreSQL   │           │  PostgreSQL   │
                                 │     :5433     │           │     :5434     │
                                 └───────────────┘           └───────────────┘
```

## Tecnologias

### Frontend

- Angular
- TypeScript
- SCSS

### Backend

- Go
- Gin
- GORM

### Banco de dados

- PostgreSQL 15

### Infraestrutura

- Docker
- Docker Compose

## Pré-requisitos

- Docker
- Go 1.21+
- Node.js
- npm

## Estrutura

```text
Korp_Teste_Dihann/
├── backend/
│   ├── estoque/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go
│   │   │   ├── database/
│   │   │   │   └── database.go
│   │   │   ├── domain/
│   │   │   │   ├── produto.go
│   │   │   │   └── produto_test.go
│   │   │   └── repository/
│   │   │       ├── produto_repository.go
│   │   │       └── produto_repository_test.go
│   │   ├── migrations/
│   │   │   ├── 000001_create_produtos.up.sql
│   │   │   └── 000001_create_produtos.down.sql
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── faturamento/
│       ├── cmd/
│       │   └── main.go
│       ├── migrations/
│       ├── go.mod
│       └── go.sum
│
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── app/
│   │   ├── index.html
│   │   └── main.ts
│   ├── angular.json
│   ├── package-lock.json
│   ├── package.json
│   └── tsconfig.json
│
├── docs/
│   └── detalhamento-tecnico.md
├── .env.example
├── .gitignore
├── docker-compose.yml
└── README.md
```

## Configuração

Os microsserviços utilizam variáveis de ambiente para configuração.

Cada serviço possui seu próprio conjunto de variáveis, utilizando prefixos para evitar conflitos:

```text
ESTOQUE_DB_*
ESTOQUE_API_PORT

FATURAMENTO_DB_*
FATURAMENTO_API_PORT
```

O projeto disponibiliza o arquivo `.env.example` como referência para configuração do ambiente local.

Em ambientes de produção, as mesmas variáveis devem ser fornecidas diretamente pela infraestrutura de execução, como Azure ou containers, sem necessidade de disponibilizar um arquivo `.env`.

### Bancos de dados

Os bancos PostgreSQL são executados localmente por meio do Docker Compose, em containers independentes, cada um associado a um microsserviço e a um volume persistente.

Na raiz do projeto:

```bash
docker compose up -d
```

Para verificar os containers:

```bash
docker compose ps
```

Para interromper os containers:

```bash
docker compose down
```

### Microsserviço de Estoque

#### Migrations

As migrations do banco de dados são versionadas por microsserviço.

Para executar as migrations do Estoque:

```bash
cd backend/estoque
migrate -path migrations -database "postgres://<ESTOQUE_DB_USER>:<ESTOQUE_DB_PASSWORD>@<ESTOQUE_DB_HOST>:<ESTOQUE_DB_PORT>/<ESTOQUE_DB_NAME>?sslmode=disable" up
```
#### Testes

Para executar os testes automatizados do Estoque:
```bash
go test ./... -v
```
Os testes do repository utilizam o PostgreSQL configurado no ambiente de desenvolvimento.

#### Execução

```bash
go run ./cmd
```

**API:** [http://localhost:8080](http://localhost:8080)

**Health Check:** [http://localhost:8080/health](http://localhost:8080/health)

### Microsserviço de Faturamento

```bash
cd backend/faturamento
go run ./cmd
```

**API:** [http://localhost:8081](http://localhost:8081)

**Health Check:** [http://localhost:8081/health](http://localhost:8081/health)

### Frontend

```bash
cd frontend
npm install
npm start
```

**Aplicação:** [http://localhost:4200](http://localhost:4200)

## Portas


| Serviço                | Porta  |
| ---------------------- | ------ |
| Angular                | `4200` |
| Estoque API            | `8080` |
| Faturamento API        | `8081` |
| PostgreSQL Estoque     | `5433` |
| PostgreSQL Faturamento | `5434` |


## Documentação Técnica

O detalhamento da arquitetura, decisões técnicas, tecnologias utilizadas e implementação da solução está disponível em [docs/detalhamento-tecnico.md](docs/detalhamento-tecnico.md).