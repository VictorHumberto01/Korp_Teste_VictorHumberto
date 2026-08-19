package command

import (
	"context"

	"usuario-service/internal/application/dto"
	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/repository"
	"usuario-service/internal/domain/valueobject"
)

type UpdateUsuarioHandler struct {
	repo repository.UsuarioRepository
}

func NewUpdateUsuarioHandler(repo repository.UsuarioRepository) *UpdateUsuarioHandler {
	return &UpdateUsuarioHandler{repo: repo}
}

func (h *UpdateUsuarioHandler) Handle(ctx context.Context, id string, req dto.UpdateUsuarioRequest) (*dto.UsuarioResponse, error) {
	u, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domainerrors.ErrUsuarioNotFound
	}

	if u.Version() != req.Version {
		return nil, domainerrors.ErrConcurrencyConflict
	}

	if req.Nome != nil {
		err := u.AtualizarNome(*req.Nome)
		if err != nil {
			return nil, err
		}
	}

	if req.Email != nil {
		emailVO, err := valueobject.NewEmail(*req.Email)
		if err != nil {
			return nil, err
		}

		existingUser, findErr := h.repo.FindByEmail(ctx, emailVO)
		if findErr != nil && findErr != domainerrors.ErrUsuarioNotFound {
			return nil, findErr
		}
		if existingUser != nil && existingUser.ID() != id {
			return nil, domainerrors.ErrEmailAlreadyExists
		}

		err = u.AtualizarEmail(*req.Email)
		if err != nil {
			return nil, err
		}
	}

	if req.Bio != nil {
		u.AtualizarBio(*req.Bio)
	}

	u.IncrementVersion()

	err = h.repo.Update(ctx, u)
	if err != nil {
		return nil, err
	}

	res := dto.FromEntity(u)
	return &res, nil
}
