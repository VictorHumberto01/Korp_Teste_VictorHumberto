package query

import (
	"context"
	"math"

	"estoque-service/internal/application/dto"
	"estoque-service/internal/domain/repository"
)

type ListProdutosHandler struct {
	repo repository.ProdutoRepository
}

func NewListProdutosHandler(repo repository.ProdutoRepository) *ListProdutosHandler {
	return &ListProdutosHandler{repo: repo}
}

func (h *ListProdutosHandler) Handle(ctx context.Context, page, pageSize int) (*dto.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	produtos, total, err := h.repo.FindAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.ProdutoResponse, 0, len(produtos))
	for _, p := range produtos {
		data = append(data, dto.FromEntity(p))
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &dto.PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
