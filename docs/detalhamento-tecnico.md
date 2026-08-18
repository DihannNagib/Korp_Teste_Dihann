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
                                 ┌───────▼───────┐           ┌───────▼───────┐
                                 │    Estoque    │           │  Faturamento  │
                                 │     :8080     │           │     :8081     │
                                 │      Go       │           │      Go       │
                                 └───────┬───────┘           └───────┬───────┘
                                         │                           │
                                 ┌───────▼───────┐           ┌───────▼───────┐
                                 │  PostgreSQL   │           │  PostgreSQL   │
                                 │     :5433     │           │     :5434     │
                                 └───────────────┘           └───────────────┘
```



### 2.1 Serviço de Estoque

O serviço de Estoque é responsável pelo domínio de produtos e pelo controle de seus saldos.

A implementação atual contempla:

- configuração por variáveis de ambiente;
- conexão com PostgreSQL utilizando GORM;
- migration da tabela `produtos`;
- entidade de domínio `Produto`;
- regra de domínio para impedir saldo negativo;
- repository para persistência de produtos;
- criação de produtos;
- busca de produto por código;
- listagem de produtos;
- ajuste de saldo;
- tratamento de código de produto duplicado;
- tratamento de produto não encontrado;
- transação para alteração de saldo;
- bloqueio pessimista com `FOR UPDATE`;
- testes unitários do domínio;
- testes do repository.

**Porta:** `8080`

**Health Check:** `GET /health`

### 2.2 Serviço de Faturamento

O serviço de Faturamento será responsável pelo domínio de notas fiscais.

**Porta:** `8081`

**Health Check:** `GET /health`

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

Os microsserviços backend são desenvolvidos em Go utilizando Gin para criação das APIs HTTP e GORM para acesso ao banco de dados.

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

Os volumes permitem persistir os dados do PostgreSQL independentemente do ciclo de vida dos containers.

A remoção dos volumes é uma operação separada e deve ser realizada explicitamente quando necessário.

---



## 5. Configuração

A aplicação utiliza variáveis de ambiente para configuração dos microsserviços.

Cada microsserviço possui seu próprio módulo de configuração, responsável por carregar e validar as variáveis necessárias.

As variáveis são prefixadas de acordo com o microsserviço:

```text
ESTOQUE_DB_*
ESTOQUE_API_PORT

FATURAMENTO_DB_*
FATURAMENTO_API_PORT
```

O projeto disponibiliza o arquivo `.env.example` como referência para configuração local.

O arquivo `.env` não deve ser versionado.

### 5.1 Desenvolvimento local

Durante o desenvolvimento, o pacote `godotenv` é utilizado para carregar as variáveis definidas no arquivo `.env`.

Fluxo:

```text
.env
 ↓
godotenv.Load()
 ↓
variáveis de ambiente
 ↓
config.Load()
 ↓
configuração da aplicação
```



### 5.2 Produção

Em produção, as variáveis podem ser fornecidas diretamente pela infraestrutura de execução, como containers ou Azure App Service.

Não é necessário disponibilizar um arquivo `.env`.

---



## 6. Organização do Backend

Cada microsserviço possui seu próprio módulo Go e gerencia suas próprias dependências.

### 6.1 Gerenciamento de Dependências

O backend utiliza Go Modules.

Cada microsserviço possui:

- `go.mod`;
- `go.sum`.



### 6.2 Conexão com Banco

Cada microsserviço possui sua própria conexão com banco de dados:

```text
Estoque
    └── internal/database
            └── PostgreSQL do Estoque

Faturamento
    └── internal/database
            └── PostgreSQL do Faturamento
```



### 6.3 Domínio de Estoque

O domínio de Estoque possui a entidade `Produto`, responsável por representar um produto e seu saldo atual.


| Campo       | Tipo        | Descrição                  |
| ----------- | ----------- | -------------------------- |
| `id`        | `uint`      | Identificador do produto   |
| `codigo`    | `string`    | Código único do produto    |
| `descricao` | `string`    | Descrição do produto       |
| `saldo`     | `int`       | Saldo atual em estoque     |
| `createdAt` | `time.Time` | Data de criação            |
| `updatedAt` | `time.Time` | Data da última atualização |


A entidade possui a regra de negócio que impede que o saldo fique negativo.

A operação de ajuste é realizada pelo método:

```text
Produto.AjustarSaldo(delta)
```

A validação pertence ao domínio e não depende do banco de dados ou da camada HTTP.

### 6.4 Persistência e Concorrência

A persistência dos produtos é realizada pelo `repository`, utilizando GORM.

O ajuste de saldo é executado dentro de uma transação e utiliza bloqueio pessimista:

```sql
SELECT ...
FROM produtos
WHERE codigo = ?
FOR UPDATE;
```

Fluxo:

```text
Início da transação
        ↓
SELECT ... FOR UPDATE
        ↓
Produto bloqueado
        ↓
Produto.AjustarSaldo(delta)
        ↓
UPDATE
        ↓
COMMIT
```

O `FOR UPDATE` impede que operações concorrentes sobre o mesmo produto trabalhem simultaneamente sobre o mesmo estado do saldo.

A regra de negócio permanece no domínio, enquanto o controle de transação, bloqueio e persistência permanece no repository.

---



## 7. API e Health Check

Cada microsserviço disponibiliza um endpoint de Health Check que verifica a disponibilidade da conexão com o PostgreSQL.

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
- Configuração do microsserviço de Estoque;
- Conexão do Estoque com PostgreSQL utilizando GORM;
- Health Check do Estoque com verificação do banco;
- Migration da tabela `produtos`;
- Domínio `Produto`;
- Regra de negócio para impedir saldo negativo;
- Repository de produtos;
- Criação e consulta de produtos;
- Ajuste de saldo;
- Tratamento de código duplicado e produto não encontrado;
- Transação e bloqueio pessimista no ajuste de saldo;
- Testes unitários do domínio;
- Testes do repository.

