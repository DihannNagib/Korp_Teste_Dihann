# Detalhamento Técnico

## 1. Visão Geral

Este documento apresenta o detalhamento técnico da solução desenvolvida para o teste técnico do Sistema de Emissão de Notas Fiscais.

A aplicação foi estruturada utilizando:

- Frontend em Angular;
- Microsserviço de Estoque desenvolvido em Go;
- Microsserviço de Faturamento desenvolvido em Go;
- Um banco PostgreSQL independente para cada microsserviço;
- Docker Compose para execução dos bancos de dados em ambiente local.

---

## 2. Arquitetura

A solução utiliza uma arquitetura baseada em microsserviços, com separação entre os domínios de Estoque e Faturamento.

```text
                                              ┌──────────────────┐
                                              │     Angular      │
                                              │      :4200       │
                                              └────────┬─────────┘
                                                        │
                                          ┌─────────────┴─────────────┐
                                          │                           │
                                  ┌──────▼───────┐            ┌──────▼───────┐
                                  │   Estoque    │            │ Faturamento  │
                                  │    :8080     │            │    :8081     │
                                  │      Go      │            │      Go      │
                                  └──────┬───────┘            └──────┬───────┘
                                          │                           │
                                  ┌──────▼───────┐            ┌──────▼───────┐
                                  │ PostgreSQL   │            │ PostgreSQL   │
                                  │    :5433     │            │    :5434     │
                                  └──────────────┘            └──────────────┘

```

## 2.1 Serviço de Estoque

O serviço de Estoque será responsável pelo domínio de produtos e controle de saldos.

- **Porta:** `8080`
- **Health Check:** `GET /health`

## 2.2 Serviço de Faturamento

O serviço de Faturamento será responsável pelo domínio de notas fiscais.

- **Porta:** `8081`
- **Health Check:** `GET /health`

---

## 3. Tecnologias

### 3.1 Frontend

- Angular
- TypeScript
- SCSS

O frontend foi inicializado em Angular e será utilizado para implementar as telas e os fluxos do sistema.

### 3.2 Backend

- Go
- Gin

Os microsserviços backend foram desenvolvidos em Go utilizando o framework Gin para criação das APIs HTTP.

### 3.3 Banco de Dados

- PostgreSQL 15

Foram configuradas duas instâncias independentes de PostgreSQL:


| Microsserviço | Banco      | Porta  |
| ------------- | ---------- | ------ |
| Estoque       | PostgreSQL | `5433` |
| Faturamento   | PostgreSQL | `5434` |


---

## 4. Infraestrutura

### 4.1 Docker Compose

O Docker Compose é utilizado para executar os bancos PostgreSQL em containers independentes.

Cada banco possui:

- container próprio;
- database próprio;
- credenciais próprias;
- porta própria;
- volume persistente próprio.

### 4.2 Persistência

Cada banco possui um Docker Named Volume próprio:

```text
estoque_data
faturamento_data
```

Os volumes são utilizados para persistir os dados do PostgreSQL fora do ciclo de vida dos containers.

Dessa forma, a remoção e recriação dos containers não implica, por padrão, na remoção dos dados armazenados nos volumes.

A remoção dos volumes é uma operação separada e deve ser realizada explicitamente quando necessário.

---

## 5. Configuração

As configurações utilizadas pelo Docker Compose são fornecidas por variáveis de ambiente.

O projeto disponibiliza um arquivo:

```text
.env.example
```

A configuração local é realizada criando um arquivo `.env` a partir do exemplo.

As credenciais e configurações sensíveis não são versionadas no repositório.

---

## 6. Organização do Backend

Cada microsserviço possui seu próprio módulo Go.

A existência de módulos independentes permite que cada microsserviço gerencie suas próprias dependências.

### 6.1 Gerenciamento de Dependências

O backend utiliza Go Modules para gerenciamento das dependências.

Cada microsserviço possui:

- `go.mod` — definição do módulo e suas dependências;
- `go.sum` — registro dos checksums das dependências.

---

## 7. API e Health Check

Cada microsserviço disponibiliza um endpoint de Health Check.

### Estoque

**Método:**

```http
GET /health
```

**Endpoint:**

```text
http://localhost:8080/health
```

### Faturamento

**Método:**

```http
GET /health
```

**Endpoint:**

```text
http://localhost:8081/health
```

O Health Check permite verificar se o processo da API está ativo e respondendo às requisições HTTP.

---

## 8. Frontend Angular

O frontend foi inicializado utilizando Angular.

O servidor de desenvolvimento utiliza a porta:

```text
4200
```

---

## Estado Atual da Implementação

### Implementado

- Estrutura inicial do projeto;
- Arquitetura inicial de microsserviços;
- Microsserviço de Estoque;
- Microsserviço de Faturamento;
- Health Check do Estoque;
- Health Check do Faturamento;
- Docker Compose;
- PostgreSQL do Estoque;
- PostgreSQL do Faturamento;
- Volumes persistentes dos bancos;
- Configuração por variáveis de ambiente;
- Go Modules;
- Angular inicializado.

