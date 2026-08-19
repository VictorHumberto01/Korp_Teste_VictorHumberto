package repository

import (
	"context"
	"usuario-service/internal/domain/entity"
	"usuario-service/internal/domain/valueobject"
)

type UsuarioRepository interface {
	Save(ctx context.Context, usuario *entity.Usuario) error
	FindByID(ctx context.Context, id string) (*entity.Usuario, error)
	FindByEmail(ctx context.Context, email valueobject.Email) (*entity.Usuario, error)
	FindByCPF(ctx context.Context, cpf valueobject.CPF) (*entity.Usuario, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*entity.Usuario, int64, error)
	Update(ctx context.Context, usuario *entity.Usuario) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error)
	ExistsByCPF(ctx context.Context, cpf valueobject.CPF) (bool, error)
}
