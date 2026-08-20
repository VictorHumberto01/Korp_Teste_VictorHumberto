package dto

type GerarDescricaoProdutoRequest struct {
	Nome string `json:"nome" binding:"required,min=1,max=200"`
}

type GerarDescricaoProdutoResponse struct {
	Descricao string `json:"descricao"`
}
