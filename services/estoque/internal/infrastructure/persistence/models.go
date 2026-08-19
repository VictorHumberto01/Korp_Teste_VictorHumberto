package persistence

import (
	"time"

	"gorm.io/gorm"

	"estoque-service/internal/domain/entity"
	"estoque-service/internal/domain/valueobject"
)

type ProdutoModel struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	Codigo    string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	Descricao string         `gorm:"type:varchar(200);not null"`
	Saldo     int            `gorm:"not null;default:0"`
	Version   int            `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ProdutoModel) TableName() string { return "produtos" }

func (m *ProdutoModel) ToDomain() (*entity.Produto, error) {
	codigo, err := valueobject.NewCodigoProduto(m.Codigo)
	if err != nil {
		return nil, err
	}

	descricao, err := valueobject.NewDescricao(m.Descricao)
	if err != nil {
		return nil, err
	}

	return entity.ReconstructProduto(m.ID, codigo, descricao, m.Saldo, m.Version, m.CreatedAt, m.UpdatedAt), nil
}

func FromDomain(p *entity.Produto) *ProdutoModel {
	return &ProdutoModel{
		ID:        p.ID(),
		Codigo:    p.Codigo().Value(),
		Descricao: p.Descricao().Value(),
		Saldo:     p.Saldo(),
		Version:   p.Version(),
		CreatedAt: p.CriadoEm(),
		UpdatedAt: p.AtualizadoEm(),
	}
}
