package command

import (
	"context"

	"estoque-service/internal/application/dto"
	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/repository"
)

type UpdateProdutoHandler struct {
	repo repository.ProdutoRepository
}

func NewUpdateProdutoHandler(repo repository.ProdutoRepository) *UpdateProdutoHandler {
	return &UpdateProdutoHandler{repo: repo}
}

func (h *UpdateProdutoHandler) Handle(ctx context.Context, id string, req dto.UpdateProdutoRequest) (*dto.ProdutoResponse, error) {
	p, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domainerrors.ErrProdutoNotFound
	}

	if p.Version() != req.Version {
		return nil, domainerrors.ErrConcurrencyConflict
	}

	if req.Descricao != nil {
		if err := p.AtualizarDescricao(*req.Descricao); err != nil {
			return nil, err
		}
	}

	p.IncrementVersion()

	if err := h.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	res := dto.FromEntity(p)
	return &res, nil
}
