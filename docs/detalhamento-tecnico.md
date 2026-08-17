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

Atualmente, o serviço possui:

- configuração por variáveis de ambiente;
- carregamento de `.env` para desenvolvimento local;
- conexão com PostgreSQL utilizando GORM;
- validação da configuração obrigatória;
- endpoint de Health Check;
- verificação da disponibilidade da conexão com o banco.
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
- GORM

Os microsserviços backend são desenvolvidos em Go utilizando o framework Gin para criação das APIs HTTP e GORM para acesso ao banco de dados.

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

A aplicação utiliza variáveis de ambiente para configuração dos microsserviços.

Cada microsserviço possui seu próprio módulo de configuração, responsável por:

- carregar as variáveis de ambiente;
- carregar o arquivo `.env` durante o desenvolvimento local;
- validar as variáveis obrigatórias;
- disponibilizar as configurações necessárias para os demais componentes do serviço.

As variáveis são prefixadas de acordo com o microsserviço para evitar conflitos de configuração.

Exemplo:

```text
ESTOQUE_DB_HOST
ESTOQUE_DB_PORT
ESTOQUE_DB_USER
ESTOQUE_DB_PASSWORD
ESTOQUE_DB_NAME
ESTOQUE_API_PORT

FATURAMENTO_DB_HOST
FATURAMENTO_DB_PORT
FATURAMENTO_DB_USER
FATURAMENTO_DB_PASSWORD
FATURAMENTO_DB_NAME
FATURAMENTO_API_PORT
```

O projeto disponibiliza um arquivo:

```
.env.example
```

Para desenvolvimento local, deve ser criado um arquivo `.env` na raiz do projeto a partir do exemplo.

O arquivo `.env` não deve ser versionado no repositório.

### 5.1 Desenvolvimento local

Durante o desenvolvimento, o pacote `godotenv` é utilizado para carregar as variáveis definidas no arquivo `.env`.

O fluxo de configuração é:

```
.env
 ↓
godotenv.Load()
 ↓
variáveis de ambiente do processo
 ↓
config.Load()
 ↓
Configuração tipada do microsserviço
```

Cada microsserviço carrega somente as variáveis necessárias ao seu próprio domínio.

### 5.2 Produção

Em produção, não é necessário disponibilizar um arquivo `.env`.

As variáveis de ambiente podem ser fornecidas diretamente pela infraestrutura de execução, como containers, Azure App Service ou outros serviços de hospedagem.

O código da aplicação não precisa ser alterado entre desenvolvimento e produção:

```
Desenvolvimento:
.env → godotenv → ambiente do processo → aplicação

Produção:
Azure/Container → ambiente do processo → aplicação
```

Dessa forma, o mesmo mecanismo de configuração baseado em `os.LookupEnv()` é utilizado nos dois ambientes.

Credenciais e demais informações sensíveis não são versionadas no repositório.

---



## 6. Organização do Backend

Cada microsserviço possui seu próprio módulo Go.

A existência de módulos independentes permite que cada microsserviço gerencie suas próprias dependências.

### 6.1 Gerenciamento de Dependências

O backend utiliza Go Modules para gerenciamento das dependências.

Cada microsserviço possui:

- `go.mod` — definição do módulo e suas dependências;
- `go.sum` — registro dos checksums das dependências.



### 6.2 Conexão com Banco

Cada microsserviço encapsula sua própria conexão com banco de dados.

A estrutura segue o princípio de isolamento por serviço:

```text
Estoque
    └── internal/database
            └── PostgreSQL do Estoque
Faturamento
    └── internal/database
            └── PostgreSQL do Faturamento
```

---



## 7. API e Health Check

Cada microsserviço disponibiliza um endpoint de Health Check.

O endpoint verifica a disponibilidade da conexão com o PostgreSQL.

Quando a aplicação e o banco estão disponíveis, retorna HTTP 200.

Quando a conexão com o banco não está disponível, retorna HTTP 503.

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
- Arquitetura baseada em microsserviços;
- Microsserviço de Estoque;
- Estrutura inicial do microsserviço de Faturamento;
- Docker Compose;
- PostgreSQL do Estoque;
- PostgreSQL do Faturamento;
- Volumes persistentes;
- Go Modules;
- Angular inicializado;
- Configuração por variáveis de ambiente;
- Carregamento de `.env` para desenvolvimento local;
- Validação das variáveis obrigatórias do Estoque;
- Configuração encapsulada no microsserviço de Estoque;
- Conexão do Estoque com PostgreSQL utilizando GORM;
- Validação da conexão com PostgreSQL;
- Health Check do Estoque com verificação do banco.

