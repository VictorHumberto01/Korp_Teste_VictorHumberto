package repository

import (
	"context"

	"faturamento-service/internal/domain/entity"
)

type NotaFiscalRepository interface {
	Save(ctx context.Context, nota *entity.NotaFiscal) error
	FindByID(ctx context.Context, id string) (*entity.NotaFiscal, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*entity.NotaFiscal, int64, error)
	Update(ctx context.Context, nota *entity.NotaFiscal) error
}
