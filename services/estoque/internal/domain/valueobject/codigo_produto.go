package valueobject

import (
	"strings"

	domainerrors "estoque-service/internal/domain/errors"
)

type CodigoProduto struct {
	value string
}

func NewCodigoProduto(value string) (CodigoProduto, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 1 || len(trimmed) > 50 {
		return CodigoProduto{}, domainerrors.ErrCodigoInvalido
	}
	return CodigoProduto{value: trimmed}, nil
}

func (c CodigoProduto) Value() string {
	return c.value
}

func (c CodigoProduto) Equals(other CodigoProduto) bool {
	return c.value == other.value
}

func (c CodigoProduto) String() string {
	return c.value
}
