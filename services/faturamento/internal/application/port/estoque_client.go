package port

import "context"

// EstoqueClient é o port de saída para comunicação com o serviço de Estoque.
// A implementação concreta (infrastructure/estoque) trata HTTP, timeouts e
// circuit breaker; a camada de aplicação só conhece esta interface.
type EstoqueClient interface {
	DebitarSaldo(ctx context.Context, produtoID string, quantidade int) error
	CreditarSaldo(ctx context.Context, produtoID string, quantidade int) error
}
