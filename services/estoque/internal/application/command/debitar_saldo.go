package command

import (
	"context"
	"errors"

	"estoque-service/internal/application/dto"
	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/repository"
)

const maxOptimisticRetries = 5

type DebitarSaldoHandler struct {
	repo repository.ProdutoRepository
}

func NewDebitarSaldoHandler(repo repository.ProdutoRepository) *DebitarSaldoHandler {
	return &DebitarSaldoHandler{repo: repo}
}

// Handle debita a quantidade do saldo do produto usando concorrência otimista.
// Em caso de conflito de versão (concorrência), relê o produto e tenta novamente.
// Se o saldo for insuficiente em qualquer tentativa, retorna erro imediatamente
// (sem retry), pois é uma regra de negócio e não um conflito de concorrência.
func (h *DebitarSaldoHandler) Handle(ctx context.Context, id string, quantidade int) (*dto.ProdutoResponse, error) {
	for i := 0; i < maxOptimisticRetries; i++ {
		p, err := h.repo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, domainerrors.ErrProdutoNotFound
		}

		if err := p.DebitarSaldo(quantidade); err != nil {
			return nil, err
		}
		p.IncrementVersion()

		err = h.repo.Update(ctx, p)
		if err == nil {
			res := dto.FromEntity(p)
			return &res, nil
		}
		if errors.Is(err, domainerrors.ErrConcurrencyConflict) {
			continue
		}
		return nil, err
	}
	return nil, domainerrors.ErrConcurrencyConflict
}
