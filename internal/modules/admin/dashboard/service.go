package dashboard

import (
	"context"
	"errors"
	"time"
)

var ErrSalaoInvalido = errors.New("id_salao inválido")

type Service struct {
	repository *Repository
	location   *time.Location
}

func NewService(repository *Repository) *Service {
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		location = time.FixedZone("BRT", -3*60*60)
	}

	return &Service{
		repository: repository,
		location:   location,
	}
}

func dataLocalAgora(location *time.Location) time.Time {
	agora := time.Now().In(location)

	return time.Date(
		agora.Year(),
		agora.Month(),
		agora.Day(),
		0, 0, 0, 0,
		location,
	)
}

func (s *Service) Resumo(
	ctx context.Context,
	idSalao int64,
) (*ResumoResponse, error) {
	if idSalao <= 0 {
		return nil, ErrSalaoInvalido
	}

	hoje := dataLocalAgora(s.location)
	amanha := hoje.AddDate(0, 0, 1)

	inicioSemana := hoje.AddDate(0, 0, -6)

	inicioMes := time.Date(
		hoje.Year(),
		hoje.Month(),
		1,
		0, 0, 0, 0,
		s.location,
	)

	inicioProximoMes := inicioMes.AddDate(0, 1, 0)

	totalHoje, err := s.repository.ContarAgendamentos(
		ctx, idSalao, hoje, amanha, false,
	)
	if err != nil {
		return nil, err
	}

	totalSemana, err := s.repository.ContarAgendamentos(
		ctx, idSalao, inicioSemana, amanha, false,
	)
	if err != nil {
		return nil, err
	}

	totalMes, err := s.repository.ContarAgendamentos(
		ctx, idSalao, inicioMes, inicioProximoMes, false,
	)
	if err != nil {
		return nil, err
	}

	canceladosHoje, err := s.repository.ContarAgendamentos(
		ctx, idSalao, hoje, amanha, true,
	)
	if err != nil {
		return nil, err
	}

	canceladosSemana, err := s.repository.ContarAgendamentos(
		ctx, idSalao, inicioSemana, amanha, true,
	)
	if err != nil {
		return nil, err
	}

	porProfissional, err := s.repository.ContarPorProfissional(
		ctx,
		idSalao,
		hoje,
		amanha,
		inicioSemana,
		amanha,
	)
	if err != nil {
		return nil, err
	}

	return &ResumoResponse{
		TotalHoje:        totalHoje,
		TotalSemana:      totalSemana,
		TotalMes:         totalMes,
		CanceladosHoje:   canceladosHoje,
		CanceladosSemana: canceladosSemana,
		PorProfissional:  porProfissional,
	}, nil
}