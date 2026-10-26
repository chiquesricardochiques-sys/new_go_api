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
		location = time.FixedZone(
			"BRT",
			-3*60*60,
		)
	}

	return &Service{
		repository: repository,
		location:   location,
	}
}

// ============================================================
// DATA LOCAL
// ============================================================

func dataLocalAgora(location *time.Location) time.Time {

	agora := time.Now().In(location)

	return time.Date(
		agora.Year(),
		agora.Month(),
		agora.Day(),
		0,
		0,
		0,
		0,
		location,
	)
}

// ============================================================
// INÍCIO DA SEMANA
// SEGUNDA-FEIRA
// ============================================================

func inicioDaSemana(data time.Time) time.Time {

	// Go:
	// Sunday = 0
	// Monday = 1
	// Tuesday = 2
	// ...
	//
	// Queremos segunda-feira como início.

	diasDesdeSegunda := (int(data.Weekday()) + 6) % 7

	return data.AddDate(
		0,
		0,
		-diasDesdeSegunda,
	)
}

// ============================================================
// RESUMO DO DASHBOARD
// ============================================================

func (s *Service) Resumo(
	ctx context.Context,
	idSalao int64,
) (*ResumoResponse, error) {

	if idSalao <= 0 {
		return nil, ErrSalaoInvalido
	}

	// --------------------------------------------------------
	// HOJE
	// --------------------------------------------------------

	hoje := dataLocalAgora(s.location)

	amanha := hoje.AddDate(
		0,
		0,
		1,
	)

	// --------------------------------------------------------
	// SEMANA
	// --------------------------------------------------------

	inicioSemana := inicioDaSemana(hoje)

	fimSemana := inicioSemana.AddDate(
		0,
		0,
		7,
	)

	// --------------------------------------------------------
	// MÊS
	// --------------------------------------------------------

	inicioMes := time.Date(
		hoje.Year(),
		hoje.Month(),
		1,
		0,
		0,
		0,
		0,
		s.location,
	)

	inicioProximoMes := inicioMes.AddDate(
		0,
		1,
		0,
	)

	// ========================================================
	// AGENDAMENTOS
	// ========================================================

	totalHoje, err := s.repository.ContarAgendamentos(
		ctx,
		idSalao,
		hoje,
		amanha,
	)

	if err != nil {
		return nil, err
	}

	totalSemana, err := s.repository.ContarAgendamentos(
		ctx,
		idSalao,
		inicioSemana,
		fimSemana,
	)

	if err != nil {
		return nil, err
	}

	totalMes, err := s.repository.ContarAgendamentos(
		ctx,
		idSalao,
		inicioMes,
		inicioProximoMes,
	)

	if err != nil {
		return nil, err
	}

	// ========================================================
	// FATURAMENTO
	// ========================================================

	faturamentoHoje, err := s.repository.CalcularFaturamento(
		ctx,
		idSalao,
		hoje,
		amanha,
	)

	if err != nil {
		return nil, err
	}

	faturamentoSemana, err := s.repository.CalcularFaturamento(
		ctx,
		idSalao,
		inicioSemana,
		fimSemana,
	)

	if err != nil {
		return nil, err
	}

	faturamentoMes, err := s.repository.CalcularFaturamento(
		ctx,
		idSalao,
		inicioMes,
		inicioProximoMes,
	)

	if err != nil {
		return nil, err
	}

	// ========================================================
	// FATURAMENTO POR MÊS
	// ========================================================

	faturamentoPorMes, err := s.repository.FaturamentoPorMes(
		ctx,
		idSalao,
	)

	if err != nil {
		return nil, err
	}

	// ========================================================
	// PROFISSIONAIS
	// ========================================================

	profissionais, err := s.repository.ContarPorProfissional(
		ctx,
		idSalao,
		hoje,
		amanha,
		inicioSemana,
		fimSemana,
	)

	if err != nil {
		return nil, err
	}

	// ========================================================
	// RESPONSE
	// ========================================================

	return &ResumoResponse{

		Agendamentos: AgendamentosResumo{
			Hoje:   totalHoje,
			Semana: totalSemana,
			Mes:    totalMes,
		},

		Faturamento: FaturamentoResumo{
			Hoje:   faturamentoHoje,
			Semana: faturamentoSemana,
			Mes:    faturamentoMes,
		},

		FaturamentoMes: faturamentoPorMes,

		Profissionais: profissionais,

	}, nil
}