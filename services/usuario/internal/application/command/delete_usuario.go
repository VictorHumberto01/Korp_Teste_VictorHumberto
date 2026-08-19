package command

import (
	"context"

	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/repository"
)

type DeleteUsuarioHandler struct {
	repo repository.UsuarioRepository
}

func NewDeleteUsuarioHandler(repo repository.UsuarioRepository) *DeleteUsuarioHandler {
	return &DeleteUsuarioHandler{repo: repo}
}

func (h *DeleteUsuarioHandler) Handle(ctx context.Context, id string) error {
	u, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if u == nil {
		return domainerrors.ErrUsuarioNotFound
	}

	err = u.Desativar()
	if err != nil {
		return err
	}

	u.IncrementVersion()

	return h.repo.Update(ctx, u)
}
