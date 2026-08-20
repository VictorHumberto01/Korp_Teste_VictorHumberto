package query

import (
	"context"
	"math"

	"faturamento-service/internal/application/dto"
	"faturamento-service/internal/domain/repository"
)

type ListNotasFiscaisHandler struct {
	repo repository.NotaFiscalRepository
}

func NewListNotasFiscaisHandler(repo repository.NotaFiscalRepository) *ListNotasFiscaisHandler {
	return &ListNotasFiscaisHandler{repo: repo}
}

func (h *ListNotasFiscaisHandler) Handle(ctx context.Context, page, pageSize int) (*dto.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	notas, total, err := h.repo.FindAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.NotaFiscalResponse, 0, len(notas))
	for _, n := range notas {
		data = append(data, dto.FromEntity(n))
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
