package command

import (
	"context"

	"github.com/google/uuid"

	"usuario-service/internal/application/dto"
	"usuario-service/internal/domain/entity"
	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/repository"
	"usuario-service/internal/domain/valueobject"
)

type CreateUsuarioHandler struct {
	repo repository.UsuarioRepository
}

func NewCreateUsuarioHandler(repo repository.UsuarioRepository) *CreateUsuarioHandler {
	return &CreateUsuarioHandler{
		repo: repo,
	}
}

func (h *CreateUsuarioHandler) Handle(ctx context.Context, req dto.CreateUsuarioRequest) (*dto.UsuarioResponse, error) {
	emailVO, err := valueobject.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	cpfVO, err := valueobject.NewCPF(req.CPF)
	if err != nil {
		return nil, err
	}

	emailExists, err := h.repo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, domainerrors.ErrEmailAlreadyExists
	}

	cpfExists, err := h.repo.ExistsByCPF(ctx, cpfVO)
	if err != nil {
		return nil, err
	}
	if cpfExists {
		return nil, domainerrors.ErrCPFAlreadyExists
	}

	id := uuid.New().String()
	u, err := entity.NewUsuario(id, req.Nome, req.Email, req.CPF)
	if err != nil {
		return nil, err
	}

	err = h.repo.Save(ctx, u)
	if err != nil {
		return nil, err
	}

	res := dto.FromEntity(u)
	return &res, nil
}
