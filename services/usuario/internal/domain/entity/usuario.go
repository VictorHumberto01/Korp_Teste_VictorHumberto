package entity

import (
	"time"
	"usuario-service/internal/domain/errors"
	"usuario-service/internal/domain/valueobject"
)

type Usuario struct {
	id           string
	nome         valueobject.Nome
	email        valueobject.Email
	cpf          valueobject.CPF
	bio          string
	ativo        bool
	version      int
	criadoEm     time.Time
	atualizadoEm time.Time
}

func NewUsuario(id, nome, email, cpf string) (*Usuario, error) {
	nomeVO, err := valueobject.NewNome(nome)
	if err != nil {
		return nil, err
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	cpfVO, err := valueobject.NewCPF(cpf)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &Usuario{
		id:           id,
		nome:         nomeVO,
		email:        emailVO,
		cpf:          cpfVO,
		ativo:        true,
		version:      1,
		criadoEm:     now,
		atualizadoEm: now,
	}, nil
}

func ReconstructUsuario(id string, nome valueobject.Nome, email valueobject.Email, cpf valueobject.CPF, bio string, ativo bool, version int, criadoEm, atualizadoEm time.Time) *Usuario {
	return &Usuario{
		id:           id,
		nome:         nome,
		email:        email,
		cpf:          cpf,
		bio:          bio,
		ativo:        ativo,
		version:      version,
		criadoEm:     criadoEm,
		atualizadoEm: atualizadoEm,
	}
}

func (u *Usuario) ID() string {
	return u.id
}

func (u *Usuario) Nome() valueobject.Nome {
	return u.nome
}

func (u *Usuario) Email() valueobject.Email {
	return u.email
}

func (u *Usuario) CPF() valueobject.CPF {
	return u.cpf
}

func (u *Usuario) Bio() string {
	return u.bio
}

func (u *Usuario) Ativo() bool {
	return u.ativo
}

func (u *Usuario) Version() int {
	return u.version
}

func (u *Usuario) CriadoEm() time.Time {
	return u.criadoEm
}

func (u *Usuario) AtualizadoEm() time.Time {
	return u.atualizadoEm
}

func (u *Usuario) AtualizarNome(nome string) error {
	nomeVO, err := valueobject.NewNome(nome)
	if err != nil {
		return err
	}
	u.nome = nomeVO
	u.atualizadoEm = time.Now()
	return nil
}

func (u *Usuario) AtualizarEmail(email string) error {
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return err
	}
	u.email = emailVO
	u.atualizadoEm = time.Now()
	return nil
}

func (u *Usuario) AtualizarBio(bio string) {
	u.bio = bio
	u.atualizadoEm = time.Now()
}

func (u *Usuario) Desativar() error {
	if !u.ativo {
		return domainerrors.ErrUsuarioInativo
	}
	u.ativo = false
	u.atualizadoEm = time.Now()
	return nil
}

func (u *Usuario) Ativar() error {
	if u.ativo {
		return domainerrors.ErrUsuarioAtivo
	}
	u.ativo = true
	u.atualizadoEm = time.Now()
	return nil
}

func (u *Usuario) IncrementVersion() {
	u.version++
}
