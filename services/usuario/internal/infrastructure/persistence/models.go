package persistence

import (
	"time"

	"gorm.io/gorm"
	"usuario-service/internal/domain/entity"
	"usuario-service/internal/domain/valueobject"
)

type UsuarioModel struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	Nome      string         `gorm:"type:varchar(100);not null"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	CPF       string         `gorm:"type:varchar(14);uniqueIndex;not null"`
	Bio       string         `gorm:"type:text"`
	Ativo     bool           `gorm:"not null;default:true"`
	Version   int            `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (UsuarioModel) TableName() string { return "usuarios" }

func (m *UsuarioModel) ToDomain() (*entity.Usuario, error) {
	nome, err := valueobject.NewNome(m.Nome)
	if err != nil {
		return nil, err
	}

	email, err := valueobject.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	cpf, err := valueobject.NewCPF(m.CPF)
	if err != nil {
		return nil, err
	}

	return entity.ReconstructUsuario(m.ID, nome, email, cpf, m.Bio, m.Ativo, m.Version, m.CreatedAt, m.UpdatedAt), nil
}

func FromDomain(u *entity.Usuario) *UsuarioModel {
	return &UsuarioModel{
		ID:        u.ID(),
		Nome:      u.Nome().Value(),
		Email:     u.Email().Value(),
		CPF:       u.CPF().Value(),
		Bio:       u.Bio(),
		Ativo:     u.Ativo(),
		Version:   u.Version(),
		CreatedAt: u.CriadoEm(),
		UpdatedAt: u.AtualizadoEm(),
	}
}
