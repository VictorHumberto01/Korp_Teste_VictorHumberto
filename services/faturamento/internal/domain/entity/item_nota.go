package entity

import (
	domainerrors "faturamento-service/internal/domain/errors"
)

type ItemNota struct {
	produtoID  string
	quantidade int
}

func NewItemNota(produtoID string, quantidade int) (ItemNota, error) {
	if produtoID == "" {
		return ItemNota{}, domainerrors.ErrProdutoIDObrigatorio
	}
	if quantidade <= 0 {
		return ItemNota{}, domainerrors.ErrQuantidadeInvalida
	}
	return ItemNota{produtoID: produtoID, quantidade: quantidade}, nil
}

func (i ItemNota) ProdutoID() string {
	return i.produtoID
}

func (i ItemNota) Quantidade() int {
	return i.quantidade
}
