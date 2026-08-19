package repository

import (
	"context"

	"estoque-service/internal/domain/entity"
	"estoque-service/internal/domain/valueobject"
)

type ProdutoRepository interface {
	Save(ctx context.Context, produto *entity.Produto) error
	FindByID(ctx context.Context, id string) (*entity.Produto, error)
	FindByCodigo(ctx context.Context, codigo valueobject.CodigoProduto) (*entity.Produto, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*entity.Produto, int64, error)
	Update(ctx context.Context, produto *entity.Produto) error
	Delete(ctx context.Context, id string) error
	ExistsByCodigo(ctx context.Context, codigo valueobject.CodigoProduto) (bool, error)
}
