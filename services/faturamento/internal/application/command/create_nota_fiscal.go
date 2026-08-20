package command

import (
	"context"

	"github.com/google/uuid"

	"faturamento-service/internal/application/dto"
	"faturamento-service/internal/domain/entity"
	"faturamento-service/internal/domain/repository"
)

type CreateNotaFiscalHandler struct {
	repo repository.NotaFiscalRepository
}

func NewCreateNotaFiscalHandler(repo repository.NotaFiscalRepository) *CreateNotaFiscalHandler {
	return &CreateNotaFiscalHandler{repo: repo}
}

// Handle cria a nota fiscal em status Aberta, sem envolver o serviço de
// Estoque: o saldo só é debitado no momento da impressão.
func (h *CreateNotaFiscalHandler) Handle(ctx context.Context, req dto.CreateNotaFiscalRequest) (*dto.NotaFiscalResponse, error) {
	itens := make([]entity.ItemNota, 0, len(req.Itens))
	for _, itemReq := range req.Itens {
		item, err := entity.NewItemNota(itemReq.ProdutoID, itemReq.Quantidade)
		if err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}

	id := uuid.New().String()
	nota, err := entity.NewNotaFiscal(id, itens)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Save(ctx, nota); err != nil {
		return nil, err
	}

	// Recarrega a nota persistida para obter o número sequencial gerado
	// pelo banco de dados (coluna serial), que a entidade em memória não
	// conhece até a confirmação da escrita.
	saved, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := dto.FromEntity(saved)
	return &res, nil
}
