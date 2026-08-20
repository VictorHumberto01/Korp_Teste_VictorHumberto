package query

import (
	"context"

	"estoque-service/internal/application/dto"
)

type AIClient interface {
	GenerateDescription(ctx context.Context, nome string) (string, error)
}

type SuggestDescriptionHandler struct {
	aiClient AIClient
}

func NewSuggestDescriptionHandler(aiClient AIClient) *SuggestDescriptionHandler {
	return &SuggestDescriptionHandler{aiClient: aiClient}
}

func (h *SuggestDescriptionHandler) Handle(ctx context.Context, req dto.SuggestDescriptionRequest) (*dto.SuggestDescriptionResponse, error) {
	descricao, err := h.aiClient.GenerateDescription(ctx, req.Nome)
	if err != nil {
		return nil, err
	}

	return &dto.SuggestDescriptionResponse{
		Descricao: descricao,
	}, nil
}
