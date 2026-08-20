# faturamento-service

Emissão de notas fiscais. É o único serviço que depende de outro (`estoque`) para completar sua operação principal — por isso concentra o padrão de circuit breaker e compensação descrito em [arquitetura.md](./arquitetura.md#consistência-entre-serviços-compensação-não-transação-distribuída).

- Código: `services/faturamento`
- Porta padrão: `8082` (`FATURAMENTO_SERVER_PORT`)
- Banco: `faturamento_db` (testes usam `faturamento_test_db`)

## Modelo de domínio

`NotaFiscal` (aggregate root, `internal/domain/entity/nota_fiscal.go`):

| Campo | Tipo | Regra |
|---|---|---|
| `id` | string (UUID) | Gerado no `create` |
| `numero` | int64 | Sequencial, gerado pelo banco (coluna `serial`) — só existe depois do primeiro `Save` |
| `status` | `ABERTA` \| `FECHADA` | Começa `ABERTA`; só pode fechar uma vez |
| `itens` | `[]ItemNota` | Não pode ser vazia |
| `version` | int | Controle de concorrência otimista |

`ItemNota` é um value object simples: `produtoID` (obrigatório) + `quantidade` (> 0).

**Uma nota nunca é criada já fechada.** `Fechar()` só é chamado depois que todo o saldo dos itens foi debitado com sucesso no `estoque` — e uma nota já `FECHADA` não pode ser fechada de novo (`ErrNotaFiscalNaoAberta`).

## Fluxo de emissão: criar → imprimir

A criação da nota **não** toca no estoque — só persiste os itens com status `ABERTA`. O débito de saldo só acontece na impressão:

```
POST /api/notas-fiscais           → cria nota (ABERTA), sem chamar o estoque
POST /api/notas-fiscais/:id/imprimir → debita saldo de cada item no estoque e fecha a nota
```

Separar os dois passos permite montar a nota (adicionar itens, revisar) antes de comprometer o estoque de verdade.

### O que acontece dentro de `/imprimir`

`internal/application/command/imprimir_nota_fiscal.go`:

1. Busca a nota; se não existir → `404`; se não estiver `ABERTA` → `409`.
2. Para cada item, chama `estoque.DebitarSaldo(produtoID, quantidade)`.
3. **Se um débito falhar no meio da lista** (ex: saldo insuficiente no 3º de 5 itens), credita de volta (`CreditarSaldo`) todos os itens já debitados com sucesso antes desse ponto, e devolve o erro original. A nota continua `ABERTA` — nada fica em estado intermediário.
4. Se todos os débitos derem certo, fecha a nota (`Fechar()`) e persiste.

Isso é uma compensação manual, não uma transação distribuída — não existe rollback atômico entre dois bancos diferentes. Se a própria reversão (passo 3) falhar (o estoque caiu bem nesse momento), o erro é só logado (`log.Printf`) porque não há ação melhor a tomar automaticamente; fica registrado para investigação manual.

## Comunicação com o estoque

`internal/infrastructure/estoque/http_client.go` implementa a porta `EstoqueClient` chamando `ESTOQUE_SERVICE_URL` via HTTP, com circuit breaker (`sony/gobreaker`, mesmas regras descritas em [arquitetura.md](./arquitetura.md#comunicação-entre-serviços-http--circuit-breaker)). Respostas HTTP do estoque são traduzidas para erros de domínio do faturamento:

| Resposta do estoque | Erro no faturamento |
|---|---|
| `404` | `ErrProdutoNaoEncontrado` |
| `409` com mensagem contendo "insuficiente" | `ErrSaldoInsuficiente` |
| `5xx`, timeout, conexão recusada, circuito aberto | `ErrEstoqueIndisponivel` |
| Qualquer outro erro | repassado com prefixo `estoque:` |

## Endpoints

Prefixo: `/api/notas-fiscais`. Escritas aceitam `Idempotency-Key`.

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/` | Lista paginada |
| `GET` | `/:id` | Busca por ID |
| `POST` | `/` | Cria nota (status `ABERTA`) |
| `POST` | `/:id/imprimir` | Debita saldo no estoque e fecha a nota |
| `GET` | `/health` | Health check |

### `POST /api/notas-fiscais`

```json
{
  "itens": [
    { "produto_id": "5b2f...", "quantidade": 2 },
    { "produto_id": "9ac1...", "quantidade": 1 }
  ]
}
```

## Erros de domínio → HTTP

| Erro | HTTP |
|---|---|
| `ErrNotaFiscalNotFound`, `ErrProdutoNaoEncontrado` | 404 |
| `ErrNotaFiscalNaoAberta`, `ErrConcurrencyConflict`, `ErrSaldoInsuficiente` | 409 |
| `ErrItensObrigatorios`, `ErrQuantidadeInvalida`, `ErrProdutoIDObrigatorio` | 422 |
| `ErrEstoqueIndisponivel` | 503 |
| Erro repassado do estoque (prefixo `estoque:`) | 502 |
| Validação de request (`binding`) | 400 |
| Outro | 500 (logado no servidor) |

## Rodando isoladamente

```bash
make run-faturamento
# ou
cd services/faturamento && go run ./cmd/main.go
```

Depende do `estoque` estar no ar para o endpoint `/imprimir` funcionar — `/` (criar/listar) funciona mesmo com o estoque fora do ar.
