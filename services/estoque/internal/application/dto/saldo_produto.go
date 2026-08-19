package dto

type SaldoRequest struct {
	Quantidade int `json:"quantidade" binding:"required,gt=0"`
}
