# estoque-service

Cadastro de produtos e controle de saldo (quantidade em estoque). É o serviço com mais tráfego interno: o `faturamento` debita/credita saldo nele a cada nota fiscal, e ele mesmo consome o serviço `ia` para sugerir descrições.

- Código: `services/estoque`
- Porta padrão: `8081` (`ESTOQUE_SERVER_PORT`)
- Banco: `estoque_db` (testes usam `estoque_test_db`)

## Modelo de domínio

`Produto` (aggregate root, `internal/domain/entity/produto.go`):

| Campo | Tipo | Regra |
|---|---|---|
| `id` | string (UUID) | Gerado no `create` |
| `codigo` | `valueobject.CodigoProduto` | 1 a 50 caracteres; único no sistema |
| `descricao` | `valueobject.Descricao` | 3 a 200 **caracteres** (contagem por rune, não por byte — importante com acentuação) |
| `saldo` | int | Nunca negativo |
| `version` | int | Controle de concorrência otimista |

`DebitarSaldo(quantidade)` e `CreditarSaldo(quantidade)` são métodos do próprio agregado: rejeitam quantidade `<= 0`, e débito rejeita saldo insuficiente (`ErrSaldoInsuficiente`) antes de mexer no valor.

## Endpoints

Prefixo: `/api/produtos`. Escritas aceitam `Idempotency-Key`.

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/` | Lista paginada |
| `GET` | `/:id` | Busca por ID |
| `POST` | `/` | Cria produto |
| `PUT` | `/:id` | Atualiza descrição (requer `version`) |
| `DELETE` | `/:id` | Remove produto (hard delete) |
| `POST` | `/:id/saldo/debitar` | Debita quantidade do saldo |
| `POST` | `/:id/saldo/creditar` | Credita quantidade ao saldo |
| `POST` | `/suggest-description` | Gera descrição via IA (proxy para o serviço `ia`) |
| `GET` | `/health` | Health check |

### `POST /api/produtos/:id/saldo/debitar` e `/creditar`

```json
{ "quantidade": 5 }
```

Esses dois endpoints são os únicos do sistema chamados **por outro serviço** (o `faturamento`, na hora de imprimir uma nota). Por isso usam uma estratégia de concorrência diferente do `PUT`: em vez de exigir `version` do chamador, eles fazem **retry automático com re-leitura** — até 5 tentativas, relendo o produto e tentando de novo a cada conflito de versão. Um erro de negócio genuíno (saldo insuficiente) não entra nesse retry: é devolvido na primeira tentativa, porque tentar de novo não vai resolver.

```go
// services/estoque/internal/application/command/debitar_saldo.go
for i := 0; i < maxOptimisticRetries; i++ {
    p, _ := h.repo.FindByID(ctx, id)
    if err := p.DebitarSaldo(quantidade); err != nil {
        return nil, err // erro de negócio: não tenta de novo
    }
    p.IncrementVersion()
    if err := h.repo.Update(ctx, p); err == nil {
        return &res, nil
    } else if errors.Is(err, domainerrors.ErrConcurrencyConflict) {
        continue // relê e tenta de novo
    }
}
```

### `POST /api/produtos/suggest-description`

```json
{ "nome": "Cadeira Gamer" }
```

```json
{ "descricao": "Conforto premium e estilo imponente: a Cadeira Gamer ergonômica..." }
```

O `estoque` não fala com nenhum provedor de IA diretamente — esse endpoint apenas repassa a chamada para o serviço `ia` (`internal/infrastructure/ia/http_client.go`), protegido por circuit breaker. Se o `ia` estiver fora do ar, o circuito abre e a resposta é `503` com mensagem `"serviço de IA indisponível"`, sem travar a requisição esperando timeout. Detalhes do fluxo completo em [ia.md](./ia.md).

## Erros de domínio → HTTP

| Erro | HTTP |
|---|---|
| `ErrProdutoNotFound` | 404 |
| `ErrCodigoAlreadyExists`, `ErrConcurrencyConflict`, `ErrSaldoInsuficiente` | 409 |
| `ErrCodigoInvalido`, `ErrDescricaoInvalida`, `ErrSaldoInvalido` | 422 |
| `ErrIAIndisponivel` (circuito aberto ou falha de rede com o serviço `ia`) | 503 |
| Erro repassado pelo serviço `ia` (prefixo `ia:` na mensagem) | 502 |
| Validação de request (`binding`) | 400 |
| Outro | 500 (logado no servidor) |

## Rodando isoladamente

```bash
make run-estoque
# ou
cd services/estoque && go run ./cmd/main.go
```

Depende do serviço `ia` estar no ar (`make run-ia`) só para o endpoint `/suggest-description` — o resto do serviço funciona normalmente sem ele.
