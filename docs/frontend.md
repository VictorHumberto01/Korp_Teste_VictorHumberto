# frontend

Aplicação Angular que consome os três microsserviços com API pública (`usuario`, `estoque`, `faturamento`) diretamente do navegador. O serviço `ia` não é chamado pelo frontend — só o `estoque` fala com ele (ver [ia.md](./ia.md)).

- Código: `frontend/`
- Porta: `4200` (dev server do Angular)
- Stack: Angular 22 (standalone components, sem `NgModule`), Angular Material, RxJS, roteamento com lazy loading

## Estrutura

```
src/app/
├── core/services/
│   └── notification.service.ts     # Wrapper do MatSnackBar (success/error/warning/info)
├── shared/components/
│   ├── navbar/                      # Barra de navegação fixa
│   └── confirm-dialog/               # Dialog de confirmação genérico e reutilizável
└── features/
    ├── dashboard/                    # Visão geral (agrega dados dos 3 serviços)
    ├── usuarios/
    ├── produtos/
    └── notas-fiscais/
```

Cada pasta em `features/` segue o mesmo padrão:

```
<feature>/
├── models/<feature>.model.ts         # Interfaces TypeScript espelhando o DTO do backend
├── services/<feature>.service.ts      # HttpClient + tratamento de erro + Idempotency-Key
├── <feature>.routes.ts                 # Rotas lazy-loaded da feature
└── components/
    ├── <feature>-list/                   # Tabela (MatTable + paginator + sort + busca)
    └── <feature>-form/                    # Formulário em MatDialog (criar/editar)
```

## Roteamento

`app.routes.ts` carrega cada feature sob demanda (`loadComponent` / `loadChildren`), então o bundle inicial não inclui código de uma tela que o usuário ainda não visitou:

```ts
{ path: 'produtos', loadChildren: () => import('./features/produtos/produtos.routes')... }
```

Rota raiz (`''`) e qualquer rota desconhecida (`**`) redirecionam para `/dashboard`.

## Como cada service HTTP funciona

Os três (`UsuarioService`, `ProdutoService`, `NotaFiscalService`) seguem o mesmo formato:

- Apontam para uma base URL diferente, configurada em `environment.ts` (`apiUrl`, `estoqueApiUrl`, `faturamentoApiUrl` — uma por microsserviço, nenhum API Gateway no meio).
- Toda escrita (`create`/`update`/`imprimir`) gera um `Idempotency-Key` (UUID v4 client-side) e manda no header — é o que o middleware de idempotência do backend espera (ver [arquitetura.md](./arquitetura.md#idempotência)).
- Erros passam por um `handleError` que desembrulha o envelope de erro do backend (`{"error": {"code": ..., "message": "..."}}`) para extrair a mensagem amigável, em vez de mostrar o status HTTP cru para o usuário.
- `GET` de listagem tem `retry(2)` — falha de rede transitória tenta de novo automaticamente antes de mostrar erro.

Não existe interceptor HTTP global (`provideHttpClient()` em `app.config.ts` não registra nenhum) — o tratamento de erro é feito serviço por serviço, dentro de cada `handleError`.

## Dashboard: agregação no cliente

`dashboard.component.ts` não chama um endpoint de agregação — ele dispara as três chamadas em paralelo com `forkJoin` (usuários, produtos, notas fiscais) e calcula os totais e o "saldo baixo" no próprio componente, sobre a primeira página de cada lista. Funciona bem no tamanho atual do projeto; se o volume de dados crescesse, isso é o primeiro candidato a virar um endpoint de agregação dedicado em algum dos serviços (ou um BFF).

## Padrões de UI reutilizados

- **Listagem**: `MatTableDataSource` + `MatPaginator` + `MatSort`, com um `FormControl` de busca com `debounceTime` para não disparar requisição a cada tecla.
- **Criar/editar**: sempre um `MatDialog` abrindo o componente `*-form`, nunca uma rota própria — mantém a lista visível atrás do formulário.
- **Confirmação de ação destrutiva**: `ConfirmDialogComponent` (`shared/components/confirm-dialog`) é genérico — recebe título, mensagem e textos de botão por `MAT_DIALOG_DATA`, devolve `true`/`false` no `close()`. Usado antes de desativar usuário, remover produto etc.
- **Feedback**: `NotificationService` centraliza o `MatSnackBar` — todo componente injeta ele em vez de instanciar snackbar próprio, garantindo posição/duração consistentes.
- **Formulário com lista dinâmica de itens**: `NotaFiscalFormComponent` é o único formulário com uma quantidade variável de campos — usa `FormArray` (um `FormGroup` de `produto_id` + `quantidade` por item), com `addItem()`/`removeItem()` alterando o array em tempo de execução.

## Concorrência otimista no formulário de edição

Os formulários de edição de usuário e produto mandam o campo `version` (recebido no `GET`) de volta no `PUT`. Se o backend responder `409` (outra pessoa editou o registro nesse meio-tempo), a mensagem de erro do backend chega direto na tela via `NotificationService.error(err.message)` — não há merge automático de conflito, o usuário precisa fechar o formulário e reabrir para pegar a versão atual.

## Geração de descrição por IA

`ProdutoFormComponent.suggestDescription()` chama `ProdutoService.suggestDescription(nome)`, que bate no `estoque` (`POST /api/produtos/suggest-description`), que por sua vez repassa para o serviço `ia`. A UI mostra um spinner (`isSuggestingDesc`) durante a chamada e populam o campo `descricao` do formulário com o texto retornado — o usuário ainda pode editar antes de salvar. Se o serviço `ia` estiver fora do ar, o erro (503, "serviço de IA indisponível") aparece como snackbar, sem travar o resto do formulário.

## Ambientes

| Arquivo | Quando é usado | Aponta para |
|---|---|---|
| `environment.development.ts` | `ng serve` (dev) | `localhost:8080/8081/8082` |
| `environment.ts` | build de produção (`ng build`) | mesmas portas — trocar aqui se o backend for hospedado em outro host |

## Rodando isoladamente

```bash
make install-frontend   # só na primeira vez
make run-frontend
# ou
cd frontend && npx ng serve
```

Precisa de pelo menos `usuario`, `estoque` e `faturamento` no ar para funcionar sem erros de rede — o `ia` é opcional (só afeta o botão "Gerar com IA").
