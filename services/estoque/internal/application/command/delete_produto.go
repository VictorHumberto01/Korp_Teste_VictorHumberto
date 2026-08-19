package command

import (
	"context"

	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/repository"
)

type DeleteProdutoHandler struct {
	repo repository.ProdutoRepository
}

func NewDeleteProdutoHandler(repo repository.ProdutoRepository) *DeleteProdutoHandler {
	return &DeleteProdutoHandler{repo: repo}
}

func (h *DeleteProdutoHandler) Handle(ctx context.Context, id string) error {
	p, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return domainerrors.ErrProdutoNotFound
	}

	return h.repo.Delete(ctx, id)
}
