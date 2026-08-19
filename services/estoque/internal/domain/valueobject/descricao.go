package valueobject

import (
	"strings"

	domainerrors "estoque-service/internal/domain/errors"
)

type Descricao struct {
	value string
}

func NewDescricao(value string) (Descricao, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 3 || len(trimmed) > 200 {
		return Descricao{}, domainerrors.ErrDescricaoInvalida
	}
	return Descricao{value: trimmed}, nil
}

func (d Descricao) Value() string {
	return d.value
}

func (d Descricao) Equals(other Descricao) bool {
	return d.value == other.value
}

func (d Descricao) String() string {
	return d.value
}
