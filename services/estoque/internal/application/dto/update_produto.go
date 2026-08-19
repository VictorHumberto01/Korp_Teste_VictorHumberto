package dto

type UpdateProdutoRequest struct {
	Descricao *string `json:"descricao,omitempty"`
	Version   int     `json:"version" binding:"required"`
}
