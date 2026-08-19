package domainerrors

import "errors"

var (
	ErrUsuarioNotFound     = errors.New("usuário não encontrado")
	ErrEmailAlreadyExists  = errors.New("email já cadastrado")
	ErrCPFAlreadyExists    = errors.New("CPF já cadastrado")
	ErrCPFInvalido         = errors.New("CPF inválido")
	ErrEmailInvalido       = errors.New("formato de email inválido")
	ErrNomeInvalido        = errors.New("nome deve ter entre 3 e 100 caracteres")
	ErrConcurrencyConflict = errors.New("conflito de concorrência: registro foi modificado por outra operação")
	ErrUsuarioInativo      = errors.New("usuário já está inativo")
	ErrUsuarioAtivo        = errors.New("usuário já está ativo")
)
