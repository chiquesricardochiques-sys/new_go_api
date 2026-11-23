package codeEnt

import "errors"

var (
	ErrCodigoObrigatorio = errors.New("código do salão obrigatório")
	ErrSalaoNaoEncontrado = errors.New("salão não encontrado")
	ErrSalaoIndisponivel  = errors.New("salão indisponível para vínculo")
)