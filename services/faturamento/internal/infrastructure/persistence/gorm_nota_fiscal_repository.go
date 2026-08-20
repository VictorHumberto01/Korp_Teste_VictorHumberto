package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"faturamento-service/internal/domain/entity"
	domainerrors "faturamento-service/internal/domain/errors"
)

type GormNotaFiscalRepository struct {
	db *gorm.DB
}

func NewGormNotaFiscalRepository(db *gorm.DB) *GormNotaFiscalRepository {
	return &GormNotaFiscalRepository{db: db}
}

func (r *GormNotaFiscalRepository) Save(ctx context.Context, n *entity.NotaFiscal) error {
	model := FromDomain(n)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GormNotaFiscalRepository) FindByID(ctx context.Context, id string) (*entity.NotaFiscal, error) {
	var model NotaFiscalModel
	if err := r.db.WithContext(ctx).Preload("Itens").First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrNotaFiscalNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *GormNotaFiscalRepository) FindAll(ctx context.Context, page, pageSize int) ([]*entity.NotaFiscal, int64, error) {
	var models []NotaFiscalModel
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&NotaFiscalModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Preload("Itens").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	notas := make([]*entity.NotaFiscal, 0, len(models))
	for _, m := range models {
		n, err := m.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		notas = append(notas, n)
	}

	return notas, total, nil
}

func (r *GormNotaFiscalRepository) Update(ctx context.Context, n *entity.NotaFiscal) error {
	model := FromDomain(n)
	result := r.db.WithContext(ctx).Model(&NotaFiscalModel{}).
		Where("id = ? AND version = ?", model.ID, model.Version-1).
		Updates(map[string]interface{}{
			"status":     model.Status,
			"version":    model.Version,
			"updated_at": model.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrConcurrencyConflict
	}
	return nil
}
