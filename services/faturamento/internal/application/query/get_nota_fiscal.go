package query

import (
	"context"

	"faturamento-service/internal/application/dto"
	domainerrors "faturamento-service/internal/domain/errors"
	"faturamento-service/internal/domain/repository"
)

type GetNotaFiscalHandler struct {
	repo repository.NotaFiscalRepository
}

func NewGetNotaFiscalHandler(repo repository.NotaFiscalRepository) *GetNotaFiscalHandler {
	return &GetNotaFiscalHandler{repo: repo}
}

func (h *GetNotaFiscalHandler) Handle(ctx context.Context, id string) (*dto.NotaFiscalResponse, error) {
	nota, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if nota == nil {
		return nil, domainerrors.ErrNotaFiscalNotFound
	}

	res := dto.FromEntity(nota)
	return &res, nil
}
