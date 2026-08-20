package query

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncarDescricao(t *testing.T) {
	tests := []struct {
		name      string
		descricao string
		max       int
	}{
		{"dentro do limite", "Cadeira gamer confortável.", 200},
		{"exatamente no limite", strings.Repeat("a", 200), 200},
		{"acima do limite sem espaços", strings.Repeat("a", 250), 200},
		{"acima do limite com palavras", strings.Repeat("palavra ", 40), 200},
		{"com acentuação", strings.Repeat("ção ", 100), 200},
		{"envolta em aspas e espaços", `  "descrição entre aspas"  `, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultado := truncarDescricao(tt.descricao, tt.max)
			if n := utf8.RuneCountInString(resultado); n > tt.max {
				t.Fatalf("resultado com %d caracteres, esperado no máximo %d: %q", n, tt.max, resultado)
			}
			if strings.HasPrefix(resultado, `"`) || strings.HasSuffix(resultado, `"`) {
				t.Fatalf("resultado não deveria ter aspas nas bordas: %q", resultado)
			}
		})
	}
}
