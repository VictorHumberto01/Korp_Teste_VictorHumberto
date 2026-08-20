package domainerrors

import "errors"

var (
	ErrProdutoNotFound     = errors.New("produto não encontrado")
	ErrCodigoAlreadyExists = errors.New("código de produto já cadastrado")
	ErrCodigoInvalido      = errors.New("código deve ter entre 1 e 50 caracteres")
	ErrDescricaoInvalida   = errors.New("descrição deve ter entre 3 e 200 caracteres")
	ErrSaldoInvalido       = errors.New("saldo/quantidade inválido")
	ErrSaldoInsuficiente   = errors.New("saldo insuficiente em estoque")
	ErrConcurrencyConflict = errors.New("conflito de concorrência: registro foi modificado por outra operação")
	ErrIAIndisponivel      = errors.New("serviço de IA indisponível")
)
