package dto

import (
	"time"
	"usuario-service/internal/domain/entity"
)

type UsuarioResponse struct {
	ID           string    `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	CPF          string    `json:"cpf"`
	Bio          string    `json:"bio"`
	Ativo        bool      `json:"ativo"`
	Version      int       `json:"version"`
	CriadoEm     time.Time `json:"created_at"`
	AtualizadoEm time.Time `json:"updated_at"`
}

func FromEntity(u *entity.Usuario) UsuarioResponse {
	return UsuarioResponse{
		ID:           u.ID(),
		Nome:         u.Nome().Value(),
		Email:        u.Email().Value(),
		CPF:          u.CPF().Value(),
		Bio:          u.Bio(),
		Ativo:        u.Ativo(),
		Version:      u.Version(),
		CriadoEm:     u.CriadoEm(),
		AtualizadoEm: u.AtualizadoEm(),
	}
}

type PaginatedResponse struct {
	Data       []UsuarioResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type SuggestBioRequest struct {
	Nome  string `json:"nome" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type SuggestBioResponse struct {
	Bio string `json:"bio"`
}
