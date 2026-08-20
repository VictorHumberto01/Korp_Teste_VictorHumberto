package query

import (
	"context"
	"math"

	"usuario-service/internal/application/dto"
	"usuario-service/internal/domain/repository"
)

type ListUsuariosHandler struct {
	repo repository.UsuarioRepository
}

func NewListUsuariosHandler(repo repository.UsuarioRepository) *ListUsuariosHandler {
	return &ListUsuariosHandler{repo: repo}
}

func (h *ListUsuariosHandler) Handle(ctx context.Context, page, pageSize int) (*dto.PaginatedResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := h.repo.FindAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.UsuarioResponse, 0, len(users))
	for _, u := range users {
		data = append(data, dto.FromEntity(u))
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
