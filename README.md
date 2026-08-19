# raKorp — Sistema de Emissão de Notas Fiscais

Teste técnico da Korp: sistema para cadastro de produtos, emissão de notas
fiscais e controle de estoque, estruturado como uma arquitetura de
microsserviços com frontend em Angular.

## Visão geral


| Módulo                                                 | Responsabilidade                          | Status             |
| ------------------------------------------------------ | ----------------------------------------- | ------------------ |
| `[backend/estoque](backend/estoque/README.md)`         | Cadastro de produtos e controle de saldo  | Completo           |
| `[backend/faturamento](backend/faturamento/README.md)` | Emissão e gestão de notas fiscais         | Em desenvolvimento |
| `[frontend](frontend/README.md)`                       | Telas Angular consumindo os dois serviços | Em desenvolvimento |


Detalhamento técnico completo (ciclos de vida, bibliotecas, tratamento de
erros, decisões de arquitetura): `[docs/detalhamento-tecnico.md](docs/detalhamento-tecnico.md)`.

## Arquitetura

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

Cada microsserviço é independente: código próprio, banco de dados próprio,
ciclo de deploy próprio. O Faturamento se comunica com o Estoque via HTTP
para consultar e baixar saldo no momento da impressão da nota.

## Pré-requisitos

- Docker e Docker Compose
- Go 1.21+
- Node.js 18+ e npm
- `[golang-migrate](https://github.com/golang-migrate/migrate)` (CLI, para rodar as migrations)

## Subindo o ambiente completo

```bash
# 1. Configurar variáveis de ambiente
cp .env.example .env

# 2. Subir os bancos de dados
docker compose up -d

# 3. Rodar cada serviço (ver README específico de cada um para detalhes e migrations)
cd backend/estoque && go run ./cmd        # :8080
cd backend/faturamento && go run ./cmd    # :8081
cd frontend && npm install && npm start   # :4200
```

Instruções detalhadas de configuração, migrations e testes de cada serviço
estão nos READMEs específicos linkados na tabela acima.

## Estrutura do repositório

```text
Korp_Teste_Dihann/
├── backend/
│   ├── estoque/        → ver backend/estoque/README.md
│   └── faturamento/     → ver backend/faturamento/README.md
├── frontend/             → ver frontend/README.md
├── docs/
│   └── detalhamento-tecnico.md
├── .env.example
├── docker-compose.yml
└── README.md             (este arquivo)
```

## Portas


| Serviço                | Porta  |
| ---------------------- | ------ |
| Angular                | `4200` |
| Estoque API            | `8080` |
| Faturamento API        | `8081` |
| PostgreSQL Estoque     | `5433` |
| PostgreSQL Faturamento | `5434` |


