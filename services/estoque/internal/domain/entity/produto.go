package entity

import (
	"time"

	domainerrors "estoque-service/internal/domain/errors"
	"estoque-service/internal/domain/valueobject"
)

type Produto struct {
	id           string
	codigo       valueobject.CodigoProduto
	descricao    valueobject.Descricao
	saldo        int
	version      int
	criadoEm     time.Time
	atualizadoEm time.Time
}

func NewProduto(id, codigo, descricao string, saldo int) (*Produto, error) {
	codigoVO, err := valueobject.NewCodigoProduto(codigo)
	if err != nil {
		return nil, err
	}

	descricaoVO, err := valueobject.NewDescricao(descricao)
	if err != nil {
		return nil, err
	}

	if saldo < 0 {
		return nil, domainerrors.ErrSaldoInvalido
	}

	now := time.Now()

	return &Produto{
		id:           id,
		codigo:       codigoVO,
		descricao:    descricaoVO,
		saldo:        saldo,
		version:      1,
		criadoEm:     now,
		atualizadoEm: now,
	}, nil
}

func ReconstructProduto(id string, codigo valueobject.CodigoProduto, descricao valueobject.Descricao, saldo, version int, criadoEm, atualizadoEm time.Time) *Produto {
	return &Produto{
		id:           id,
		codigo:       codigo,
		descricao:    descricao,
		saldo:        saldo,
		version:      version,
		criadoEm:     criadoEm,
		atualizadoEm: atualizadoEm,
	}
}

func (p *Produto) ID() string {
	return p.id
}

func (p *Produto) Codigo() valueobject.CodigoProduto {
	return p.codigo
}

func (p *Produto) Descricao() valueobject.Descricao {
	return p.descricao
}

func (p *Produto) Saldo() int {
	return p.saldo
}

func (p *Produto) Version() int {
	return p.version
}

func (p *Produto) CriadoEm() time.Time {
	return p.criadoEm
}

func (p *Produto) AtualizadoEm() time.Time {
	return p.atualizadoEm
}

func (p *Produto) AtualizarDescricao(descricao string) error {
	descricaoVO, err := valueobject.NewDescricao(descricao)
	if err != nil {
		return err
	}
	p.descricao = descricaoVO
	p.atualizadoEm = time.Now()
	return nil
}

func (p *Produto) DebitarSaldo(quantidade int) error {
	if quantidade <= 0 {
		return domainerrors.ErrSaldoInvalido
	}
	if p.saldo < quantidade {
		return domainerrors.ErrSaldoInsuficiente
	}
	p.saldo -= quantidade
	p.atualizadoEm = time.Now()
	return nil
}

func (p *Produto) CreditarSaldo(quantidade int) error {
	if quantidade <= 0 {
		return domainerrors.ErrSaldoInvalido
	}
	p.saldo += quantidade
	p.atualizadoEm = time.Now()
	return nil
}

func (p *Produto) IncrementVersion() {
	p.version++
}
