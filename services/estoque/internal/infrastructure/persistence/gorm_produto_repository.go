package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"estoque-service/internal/domain/entity"
	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/valueobject"
)

type GormProdutoRepository struct {
	db *gorm.DB
}

func NewGormProdutoRepository(db *gorm.DB) *GormProdutoRepository {
	return &GormProdutoRepository{db: db}
}

func (r *GormProdutoRepository) Save(ctx context.Context, p *entity.Produto) error {
	model := FromDomain(p)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GormProdutoRepository) FindByID(ctx context.Context, id string) (*entity.Produto, error) {
	var model ProdutoModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrProdutoNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *GormProdutoRepository) FindByCodigo(ctx context.Context, codigo valueobject.CodigoProduto) (*entity.Produto, error) {
	var model ProdutoModel
	if err := r.db.WithContext(ctx).First(&model, "codigo = ?", codigo.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrProdutoNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *GormProdutoRepository) FindAll(ctx context.Context, page, pageSize int) ([]*entity.Produto, int64, error) {
	var models []ProdutoModel
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&ProdutoModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	produtos := make([]*entity.Produto, 0, len(models))
	for _, m := range models {
		p, err := m.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		produtos = append(produtos, p)
	}

	return produtos, total, nil
}

func (r *GormProdutoRepository) Update(ctx context.Context, p *entity.Produto) error {
	model := FromDomain(p)
	result := r.db.WithContext(ctx).Model(&ProdutoModel{}).
		Where("id = ? AND version = ?", model.ID, model.Version-1).
		Updates(map[string]interface{}{
			"descricao":  model.Descricao,
			"saldo":      model.Saldo,
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

func (r *GormProdutoRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&ProdutoModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrProdutoNotFound
	}
	return nil
}

func (r *GormProdutoRepository) ExistsByCodigo(ctx context.Context, codigo valueobject.CodigoProduto) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProdutoModel{}).Where("codigo = ?", codigo.Value()).Count(&count).Error
	return count > 0, err
}
