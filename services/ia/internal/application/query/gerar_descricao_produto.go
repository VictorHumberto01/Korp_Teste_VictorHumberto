package query

import (
	"context"
	"strings"

	"ia-service/internal/application/dto"
)

// maxDescricaoLength espelha o limite de caracteres do campo descrição de
// produto no estoque-service. O prompt já pede um texto curto, mas o modelo
// não obedece esse limite de forma confiável, então garantimos aqui.
const maxDescricaoLength = 200

type DescriptionProvider interface {
	GenerateDescription(ctx context.Context, nome string) (string, error)
}

type GerarDescricaoProdutoHandler struct {
	provider DescriptionProvider
}

func NewGerarDescricaoProdutoHandler(provider DescriptionProvider) *GerarDescricaoProdutoHandler {
	return &GerarDescricaoProdutoHandler{provider: provider}
}

func (h *GerarDescricaoProdutoHandler) Handle(ctx context.Context, req dto.GerarDescricaoProdutoRequest) (*dto.GerarDescricaoProdutoResponse, error) {
	descricao, err := h.provider.GenerateDescription(ctx, req.Nome)
	if err != nil {
		return nil, err
	}

	return &dto.GerarDescricaoProdutoResponse{
		Descricao: truncarDescricao(descricao, maxDescricaoLength),
	}, nil
}

func truncarDescricao(descricao string, max int) string {
	descricao = strings.Trim(strings.TrimSpace(descricao), `"'`)
	runes := []rune(descricao)
	if len(runes) <= max {
		return descricao
	}

	cortado := string(runes[:max])
	if idx := strings.LastIndexAny(cortado, " \n\t"); idx > 0 {
		cortado = cortado[:idx]
	}
	return strings.TrimRight(cortado, " ,.;:-")
}
