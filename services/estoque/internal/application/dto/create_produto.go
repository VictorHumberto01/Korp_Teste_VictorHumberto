package dto

type CreateProdutoRequest struct {
	Codigo    string `json:"codigo" binding:"required"`
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo" binding:"gte=0"`
}
