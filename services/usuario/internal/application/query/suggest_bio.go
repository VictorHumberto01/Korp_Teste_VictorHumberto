package query

import (
	"context"

	"usuario-service/internal/application/dto"
)

type AIClient interface {
	GenerateBio(ctx context.Context, nome, email string) (string, error)
}

type SuggestBioHandler struct {
	aiClient AIClient
}

func NewSuggestBioHandler(aiClient AIClient) *SuggestBioHandler {
	return &SuggestBioHandler{aiClient: aiClient}
}

func (h *SuggestBioHandler) Handle(ctx context.Context, req dto.SuggestBioRequest) (*dto.SuggestBioResponse, error) {
	bio, err := h.aiClient.GenerateBio(ctx, req.Nome, req.Email)
	if err != nil {
		return nil, err
	}

	return &dto.SuggestBioResponse{
		Bio: bio,
	}, nil
}
