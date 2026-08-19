package command

import (
	"context"
	"errors"

	"estoque-service/internal/application/dto"
	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/repository"
)

type CreditarSaldoHandler struct {
	repo repository.ProdutoRepository
}

func NewCreditarSaldoHandler(repo repository.ProdutoRepository) *CreditarSaldoHandler {
	return &CreditarSaldoHandler{repo: repo}
}

// Handle credita (devolve) a quantidade ao saldo do produto. Usado, por exemplo,
// para rollback quando o serviço de Faturamento precisa reverter um débito parcial.
func (h *CreditarSaldoHandler) Handle(ctx context.Context, id string, quantidade int) (*dto.ProdutoResponse, error) {
	for i := 0; i < maxOptimisticRetries; i++ {
		p, err := h.repo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, domainerrors.ErrProdutoNotFound
		}

		if err := p.CreditarSaldo(quantidade); err != nil {
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
