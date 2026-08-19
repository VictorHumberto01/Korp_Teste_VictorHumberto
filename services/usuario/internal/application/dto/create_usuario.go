package dto

type CreateUsuarioRequest struct {
	Nome  string `json:"nome" binding:"required"`
	Email string `json:"email" binding:"required"`
	CPF   string `json:"cpf" binding:"required"`
}
