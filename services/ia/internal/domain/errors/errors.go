package domainerrors

import "errors"

var (
	ErrNomeInvalido   = errors.New("nome deve ter entre 1 e 200 caracteres")
	ErrProviderFalhou = errors.New("provedor de IA falhou ao gerar conteúdo")
)
