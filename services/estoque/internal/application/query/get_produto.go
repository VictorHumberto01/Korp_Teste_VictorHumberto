package query

import (
	"context"

	"estoque-service/internal/application/dto"
	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/repository"
)

type GetProdutoHandler struct {
	repo repository.ProdutoRepository
}

func NewGetProdutoHandler(repo repository.ProdutoRepository) *GetProdutoHandler {
	return &GetProdutoHandler{repo: repo}
}

func (h *GetProdutoHandler) Handle(ctx context.Context, id string) (*dto.ProdutoResponse, error) {
	p, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domainerrors.ErrProdutoNotFound
	}

	res := dto.FromEntity(p)
	return &res, nil
}
