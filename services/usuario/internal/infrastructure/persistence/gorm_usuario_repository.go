package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"usuario-service/internal/domain/entity"
	domainerrors "usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/valueobject"
)

type GormUsuarioRepository struct {
	db *gorm.DB
}

func NewGormUsuarioRepository(db *gorm.DB) *GormUsuarioRepository {
	return &GormUsuarioRepository{db: db}
}

func (r *GormUsuarioRepository) Save(ctx context.Context, u *entity.Usuario) error {
	model := FromDomain(u)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GormUsuarioRepository) FindByID(ctx context.Context, id string) (*entity.Usuario, error) {
	var model UsuarioModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrUsuarioNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *GormUsuarioRepository) FindByEmail(ctx context.Context, email valueobject.Email) (*entity.Usuario, error) {
	var model UsuarioModel
	if err := r.db.WithContext(ctx).First(&model, "email = ?", email.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrUsuarioNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *GormUsuarioRepository) FindByCPF(ctx context.Context, cpf valueobject.CPF) (*entity.Usuario, error) {
	var model UsuarioModel
	if err := r.db.WithContext(ctx).First(&model, "cpf = ?", cpf.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrUsuarioNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *GormUsuarioRepository) FindAll(ctx context.Context, page, pageSize int) ([]*entity.Usuario, int64, error) {
	var models []UsuarioModel
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).Model(&UsuarioModel{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	usuarios := make([]*entity.Usuario, 0, len(models))
	for _, m := range models {
		u, err := m.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		usuarios = append(usuarios, u)
	}

	return usuarios, total, nil
}

func (r *GormUsuarioRepository) Update(ctx context.Context, u *entity.Usuario) error {
	model := FromDomain(u)
	result := r.db.WithContext(ctx).Model(&UsuarioModel{}).
		Where("id = ? AND version = ?", model.ID, model.Version-1).
		Updates(map[string]interface{}{
			"nome":       model.Nome,
			"email":      model.Email,
			"cpf":        model.CPF,
			"bio":        model.Bio,
			"ativo":      model.Ativo,
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

func (r *GormUsuarioRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&UsuarioModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrUsuarioNotFound
	}
	return nil
}

func (r *GormUsuarioRepository) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UsuarioModel{}).Where("email = ?", email.Value()).Count(&count).Error
	return count > 0, err
}

func (r *GormUsuarioRepository) ExistsByCPF(ctx context.Context, cpf valueobject.CPF) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UsuarioModel{}).Where("cpf = ?", cpf.Value()).Count(&count).Error
	return count > 0, err
}
