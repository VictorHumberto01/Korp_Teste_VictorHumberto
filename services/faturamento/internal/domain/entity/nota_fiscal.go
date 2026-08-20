package entity

import (
	"time"

	domainerrors "faturamento-service/internal/domain/errors"
)

type StatusNota string

const (
	StatusAberta  StatusNota = "ABERTA"
	StatusFechada StatusNota = "FECHADA"
)

type NotaFiscal struct {
	id           string
	numero       int64
	status       StatusNota
	itens        []ItemNota
	version      int
	criadoEm     time.Time
	atualizadoEm time.Time
}

func NewNotaFiscal(id string, itens []ItemNota) (*NotaFiscal, error) {
	if len(itens) == 0 {
		return nil, domainerrors.ErrItensObrigatorios
	}

	now := time.Now()

	return &NotaFiscal{
		id:           id,
		status:       StatusAberta,
		itens:        itens,
		version:      1,
		criadoEm:     now,
		atualizadoEm: now,
	}, nil
}

func ReconstructNotaFiscal(id string, numero int64, status StatusNota, itens []ItemNota, version int, criadoEm, atualizadoEm time.Time) *NotaFiscal {
	return &NotaFiscal{
		id:           id,
		numero:       numero,
		status:       status,
		itens:        itens,
		version:      version,
		criadoEm:     criadoEm,
		atualizadoEm: atualizadoEm,
	}
}

func (n *NotaFiscal) ID() string {
	return n.id
}

func (n *NotaFiscal) Numero() int64 {
	return n.numero
}

func (n *NotaFiscal) Status() StatusNota {
	return n.status
}

func (n *NotaFiscal) Itens() []ItemNota {
	return n.itens
}

func (n *NotaFiscal) Version() int {
	return n.version
}

func (n *NotaFiscal) CriadoEm() time.Time {
	return n.criadoEm
}

func (n *NotaFiscal) AtualizadoEm() time.Time {
	return n.atualizadoEm
}

func (n *NotaFiscal) EstaAberta() bool {
	return n.status == StatusAberta
}

// Fechar transiciona a nota para Fechada. Deve ser chamado somente após o
// débito de saldo de todos os itens ter sido confirmado no serviço de Estoque.
func (n *NotaFiscal) Fechar() error {
	if !n.EstaAberta() {
		return domainerrors.ErrNotaFiscalNaoAberta
	}
	n.status = StatusFechada
	n.atualizadoEm = time.Now()
	return nil
}

func (n *NotaFiscal) IncrementVersion() {
	n.version++
}
