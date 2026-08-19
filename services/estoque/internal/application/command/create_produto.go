package command

import (
	"context"

	"github.com/google/uuid"

	"estoque-service/internal/application/dto"
	"estoque-service/internal/domain/entity"
	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/repository"
	"estoque-service/internal/domain/valueobject"
)

type CreateProdutoHandler struct {
	repo repository.ProdutoRepository
}

func NewCreateProdutoHandler(repo repository.ProdutoRepository) *CreateProdutoHandler {
	return &CreateProdutoHandler{repo: repo}
}

func (h *CreateProdutoHandler) Handle(ctx context.Context, req dto.CreateProdutoRequest) (*dto.ProdutoResponse, error) {
	codigoVO, err := valueobject.NewCodigoProduto(req.Codigo)
	if err != nil {
		return nil, err
	}

	codigoExists, err := h.repo.ExistsByCodigo(ctx, codigoVO)
	if err != nil {
		return nil, err
	}
	if codigoExists {
		return nil, domainerrors.ErrCodigoAlreadyExists
	}

	id := uuid.New().String()
	p, err := entity.NewProduto(id, req.Codigo, req.Descricao, req.Saldo)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Save(ctx, p); err != nil {
		return nil, err
	}

	res := dto.FromEntity(p)
	return &res, nil
}
