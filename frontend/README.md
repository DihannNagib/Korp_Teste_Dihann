# Frontend — Korp Notas Fiscais

Aplicação Angular que consome os serviços de Estoque (`:8080`) e
Faturamento (`:8081`).

## Stack

- Angular (standalone components, lazy loading por rota via `loadComponent`)
- Angular Material — componentes visuais
- Reactive Forms (`@angular/forms`) — formulários e `FormArray`
- RxJS — `HttpClient`, tratamento de erro, notificação entre telas

## Rodando

Com Estoque e Faturamento já rodando:

```bash
cd frontend
npm install
npm start
```

Aplicação em http://localhost:4200.

## Configuração de ambiente

`src/environments/environment.ts` centraliza as URLs dos dois
microsserviços, evitando hardcode espalhado pelos services:

```typescript
export const environment = {
  apiUrlEstoque: 'http://localhost:8080/api/v1',
  apiUrlFaturamento: 'http://localhost:8081/api/v1',
};
```

## Estrutura

```text
src/app/
├── core/
│   ├── models/        — interfaces TypeScript (contratos de API)
│   ├── services/       — ProdutoService, NotaFiscalService
│   └── interceptors/   — log centralizado de falha de rede
└── features/
    ├── produtos/
    │   ├── produto-list/   — listagem
    │   └── produto-form/   — cadastro
    └── notas/
        ├── nota-list/      — listagem
        ├── nota-form/      — criação com múltiplos itens (FormArray + autocomplete de produtos)
        └── nota-detail/    — visualização + impressão
```

## Layout

`AppComponent` usa `mat-toolbar` + `mat-sidenav` (modo `over`) para
navegação responsiva: em telas estreitas, o menu abre como drawer
lateral via botão de hambúrguer; em telas largas, os links ficam
visíveis diretamente na toolbar.

## Validações no formulário de nota fiscal

Além da validação estrutural do Reactive Forms (`Validators.required`,
`Validators.min`), o `NotaFormComponent` carrega a lista de produtos
do Estoque (`ngOnInit`) e faz três checagens client-side antes de
enviar a requisição:

1. **Produto duplicado na mesma nota** — impede adicionar o mesmo
   código duas vezes como itens separados.
2. **Produto inexistente** — compara o código digitado/selecionado
   contra a lista de produtos carregada do Estoque.
3. **Saldo insuficiente** — soma as quantidades por produto e compara
   contra o saldo conhecido no momento do carregamento.

Essas validações são **feedback antecipado de UX**, não a fonte da
verdade: a validação definitiva continua sendo feita pelo backend no
momento da impressão, já que o saldo pode mudar entre o carregamento
da tela e o envio do formulário (ex: outra nota impressa nesse
intervalo). O autocomplete (`MatAutocompleteModule`) usa essa mesma
lista para sugerir códigos existentes durante a digitação.

## Testes

### Automatizados

Não implementados nesta versão. A cobertura de testes automatizados
do projeto está concentrada no backend (Estoque e Faturamento — ver
READMEs correspondentes); o frontend foi validado via roteiro manual
completo abaixo.

### Roteiro de teste manual completo

1. Cadastrar produto em `/produtos/novo`.
2. Criar nota com esse produto em `/notas/nova` (múltiplos itens,
   autocomplete sugerindo códigos existentes).
3. Tentar adicionar o mesmo produto duas vezes na mesma nota — erro
   client-side antes de enviar.
4. Tentar criar nota com quantidade acima do saldo disponível — erro
   client-side antes de enviar.
5. Conferir que o saldo do produto **não muda** só por criar a nota.
6. Abrir o detalhe da nota, clicar "Imprimir" — spinner aparece,
   status muda para `FECHADA`, botão de impressão some.
7. Conferir que o saldo do produto caiu a quantidade correta.
8. Tentar reimprimir a mesma nota — opção não disponível na UI (nota
   já fechada).
9. Derrubar o Estoque, criar nova nota e tentar imprimir — mensagem
   de indisponibilidade aparece, nota permanece `ABERTA`.
10. Subir o Estoque novamente, reimprimir a mesma nota — sucesso, sem
    dupla baixa de saldo.

Detalhamento das decisões técnicas:
[`../docs/detalhamento-tecnico.md`](../docs/detalhamento-tecnico.md).