package domainerrors

import "errors"

var (
	ErrNotaFiscalNotFound   = errors.New("nota fiscal não encontrada")
	ErrNotaFiscalNaoAberta  = errors.New("nota fiscal não está com status Aberta")
	ErrItensObrigatorios    = errors.New("a nota fiscal deve conter ao menos um item")
	ErrQuantidadeInvalida   = errors.New("quantidade deve ser maior que zero")
	ErrProdutoIDObrigatorio = errors.New("produto_id é obrigatório")
	ErrConcurrencyConflict  = errors.New("conflito de concorrência: registro foi modificado por outra operação")

	// Erros de comunicação com o serviço de Estoque
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado no estoque")
	ErrSaldoInsuficiente    = errors.New("saldo insuficiente em estoque")
	ErrEstoqueIndisponivel  = errors.New("serviço de estoque indisponível no momento")
)
