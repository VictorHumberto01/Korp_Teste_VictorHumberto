package dto

import (
	"time"

	"faturamento-service/internal/domain/entity"
)

type ItemNotaResponse struct {
	ProdutoID  string `json:"produto_id"`
	Quantidade int    `json:"quantidade"`
}

type NotaFiscalResponse struct {
	ID           string             `json:"id"`
	Numero       int64              `json:"numero"`
	Status       string             `json:"status"`
	Itens        []ItemNotaResponse `json:"itens"`
	Version      int                `json:"version"`
	CriadoEm     time.Time          `json:"created_at"`
	AtualizadoEm time.Time          `json:"updated_at"`
}

func FromEntity(n *entity.NotaFiscal) NotaFiscalResponse {
	itens := make([]ItemNotaResponse, 0, len(n.Itens()))
	for _, item := range n.Itens() {
		itens = append(itens, ItemNotaResponse{
			ProdutoID:  item.ProdutoID(),
			Quantidade: item.Quantidade(),
		})
	}

	return NotaFiscalResponse{
		ID:           n.ID(),
		Numero:       n.Numero(),
		Status:       string(n.Status()),
		Itens:        itens,
		Version:      n.Version(),
		CriadoEm:     n.CriadoEm(),
		AtualizadoEm: n.AtualizadoEm(),
	}
}

type PaginatedResponse struct {
	Data       []NotaFiscalResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}
