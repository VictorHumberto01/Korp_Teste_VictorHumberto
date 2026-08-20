package dto

type ItemNotaRequest struct {
	ProdutoID  string `json:"produto_id" binding:"required"`
	Quantidade int    `json:"quantidade" binding:"required,gt=0"`
}

type CreateNotaFiscalRequest struct {
	Itens []ItemNotaRequest `json:"itens" binding:"required,min=1,dive"`
}
