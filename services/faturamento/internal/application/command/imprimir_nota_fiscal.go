package command

import (
	"context"
	"log"

	"faturamento-service/internal/application/dto"
	"faturamento-service/internal/application/port"
	"faturamento-service/internal/domain/entity"
	domainerrors "faturamento-service/internal/domain/errors"
	"faturamento-service/internal/domain/repository"
)

type ImprimirNotaFiscalHandler struct {
	repo    repository.NotaFiscalRepository
	estoque port.EstoqueClient
}

func NewImprimirNotaFiscalHandler(repo repository.NotaFiscalRepository, estoque port.EstoqueClient) *ImprimirNotaFiscalHandler {
	return &ImprimirNotaFiscalHandler{repo: repo, estoque: estoque}
}

// Handle debita o saldo de cada item da nota no serviço de Estoque e, se
// todos os débitos forem bem-sucedidos, fecha a nota. Caso algum débito
// falhe no meio do caminho (produto sem saldo, serviço de Estoque
// indisponível etc.), reverte (credita de volta) os débitos já realizados
// e retorna o erro original ao usuário — a nota permanece Aberta.
func (h *ImprimirNotaFiscalHandler) Handle(ctx context.Context, id string) (*dto.NotaFiscalResponse, error) {
	nota, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if nota == nil {
		return nil, domainerrors.ErrNotaFiscalNotFound
	}
	if !nota.EstaAberta() {
		return nil, domainerrors.ErrNotaFiscalNaoAberta
	}

	debitados := make([]entity.ItemNota, 0, len(nota.Itens()))
	for _, item := range nota.Itens() {
		if err := h.estoque.DebitarSaldo(ctx, item.ProdutoID(), item.Quantidade()); err != nil {
			h.reverterDebitos(ctx, debitados)
			return nil, err
		}
		debitados = append(debitados, item)
	}

	if err := nota.Fechar(); err != nil {
		h.reverterDebitos(ctx, debitados)
		return nil, err
	}
	nota.IncrementVersion()

	if err := h.repo.Update(ctx, nota); err != nil {
		h.reverterDebitos(ctx, debitados)
		return nil, err
	}

	res := dto.FromEntity(nota)
	return &res, nil
}

// reverterDebitos credita de volta os itens já debitados com sucesso quando
// a impressão falha no meio do caminho. Falhas na reversão são apenas
// logadas: o erro original da impressão é o que importa para o usuário, e
// não há uma ação melhor a tomar aqui além de registrar a inconsistência.
func (h *ImprimirNotaFiscalHandler) reverterDebitos(ctx context.Context, debitados []entity.ItemNota) {
	for _, item := range debitados {
		if err := h.estoque.CreditarSaldo(ctx, item.ProdutoID(), item.Quantidade()); err != nil {
			log.Printf("falha ao reverter débito do produto %s (quantidade %d): %v", item.ProdutoID(), item.Quantidade(), err)
		}
	}
}
