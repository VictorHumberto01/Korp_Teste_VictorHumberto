package dto

import (
	"time"

	"estoque-service/internal/domain/entity"
)

type ProdutoResponse struct {
	ID           string    `json:"id"`
	Codigo       string    `json:"codigo"`
	Descricao    string    `json:"descricao"`
	Saldo        int       `json:"saldo"`
	Version      int       `json:"version"`
	CriadoEm     time.Time `json:"created_at"`
	AtualizadoEm time.Time `json:"updated_at"`
}

func FromEntity(p *entity.Produto) ProdutoResponse {
	return ProdutoResponse{
		ID:           p.ID(),
		Codigo:       p.Codigo().Value(),
		Descricao:    p.Descricao().Value(),
		Saldo:        p.Saldo(),
		Version:      p.Version(),
		CriadoEm:     p.CriadoEm(),
		AtualizadoEm: p.AtualizadoEm(),
	}
}

type PaginatedResponse struct {
	Data       []ProdutoResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type SuggestDescriptionRequest struct {
	Nome string `json:"nome" binding:"required"`
}

type SuggestDescriptionResponse struct {
	Descricao string `json:"descricao"`
}
