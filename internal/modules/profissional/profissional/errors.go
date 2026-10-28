package profissional

import "errors"

var (
	ErrDadosInvalidos = errors.New("dados inválidos")
	ErrNaoEncontrado  = errors.New("vínculo não encontrado")
	ErrDuplicado      = errors.New("vínculo já cadastrado")
	ErrSalaoInvalido  = errors.New("salão não encontrado")
)