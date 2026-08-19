package valueobject

import (
	"strings"
	"usuario-service/internal/domain/errors"
)

type Nome struct {
	value string
}

func NewNome(value string) (Nome, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 3 || len(trimmed) > 100 {
		return Nome{}, domainerrors.ErrNomeInvalido
	}
	return Nome{value: trimmed}, nil
}

func (n Nome) Value() string {
	return n.value
}

func (n Nome) Equals(other Nome) bool {
	return n.value == other.value
}

func (n Nome) String() string {
	return n.value
}
