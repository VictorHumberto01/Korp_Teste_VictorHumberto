package persistence

import (
	"time"

	"github.com/google/uuid"

	"faturamento-service/internal/domain/entity"
)

type NotaFiscalModel struct {
	ID        string          `gorm:"type:uuid;primaryKey"`
	Numero    int64           `gorm:"autoIncrement;unique;not null;->"`
	Status    string          `gorm:"type:varchar(20);not null"`
	Version   int             `gorm:"not null;default:1"`
	Itens     []ItemNotaModel `gorm:"foreignKey:NotaFiscalID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (NotaFiscalModel) TableName() string { return "notas_fiscais" }

type ItemNotaModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	NotaFiscalID string `gorm:"type:uuid;not null;index"`
	ProdutoID    string `gorm:"type:uuid;not null"`
	Quantidade   int    `gorm:"not null"`
}

func (ItemNotaModel) TableName() string { return "itens_nota" }

func (m *NotaFiscalModel) ToDomain() (*entity.NotaFiscal, error) {
	itens := make([]entity.ItemNota, 0, len(m.Itens))
	for _, itemModel := range m.Itens {
		item, err := entity.NewItemNota(itemModel.ProdutoID, itemModel.Quantidade)
		if err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}

	return entity.ReconstructNotaFiscal(m.ID, m.Numero, entity.StatusNota(m.Status), itens, m.Version, m.CreatedAt, m.UpdatedAt), nil
}

func FromDomain(n *entity.NotaFiscal) *NotaFiscalModel {
	itens := make([]ItemNotaModel, 0, len(n.Itens()))
	for _, item := range n.Itens() {
		itens = append(itens, ItemNotaModel{
			ID:           uuid.New().String(),
			NotaFiscalID: n.ID(),
			ProdutoID:    item.ProdutoID(),
			Quantidade:   item.Quantidade(),
		})
	}

	return &NotaFiscalModel{
		ID:        n.ID(),
		Numero:    n.Numero(),
		Status:    string(n.Status()),
		Version:   n.Version(),
		Itens:     itens,
		CreatedAt: n.CriadoEm(),
		UpdatedAt: n.AtualizadoEm(),
	}
}
