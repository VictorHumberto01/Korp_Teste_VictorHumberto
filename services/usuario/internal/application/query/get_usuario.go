package query

import (
	"context"

	"usuario-service/internal/application/dto"
	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/repository"
)

type GetUsuarioHandler struct {
	repo repository.UsuarioRepository
}

func NewGetUsuarioHandler(repo repository.UsuarioRepository) *GetUsuarioHandler {
	return &GetUsuarioHandler{repo: repo}
}

func (h *GetUsuarioHandler) Handle(ctx context.Context, id string) (*dto.UsuarioResponse, error) {
	u, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domainerrors.ErrUsuarioNotFound
	}

	res := dto.FromEntity(u)
	return &res, nil
}
