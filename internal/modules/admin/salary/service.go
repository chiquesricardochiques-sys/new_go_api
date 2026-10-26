package salary

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrInvalidData = errors.New("dados inválidos")

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Rendimento(
	ctx context.Context,
	req Request,
) (ResponseData, error) {

	if req.SalaoID <= 0 || req.ProfissionalID <= 0 {
		return ResponseData{}, ErrInvalidData
	}

	mesFiltro, err := time.ParseInLocation(
		"2006-01",
		req.MesAno,
		time.Local,
	)
	if err != nil {
		return ResponseData{}, ErrInvalidData
	}

	if mesFiltro.Year() < 2000 || mesFiltro.Year() > 2200 {
		return ResponseData{}, ErrInvalidData
	}

	profissional, err := s.repository.BuscarProfissional(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
	)
	if err != nil {
		return ResponseData{}, err
	}

	agora := time.Now()

	// ---------------------------------------------------------
	// HOJE
	// ---------------------------------------------------------

	inicioHoje := time.Date(
		agora.Year(),
		agora.Month(),
		agora.Day(),
		0,
		0,
		0,
		0,
		time.Local,
	)

	fimHoje := inicioHoje.AddDate(0, 0, 1)

	// ---------------------------------------------------------
	// SEMANA
	// Segunda-feira 00:00 -> próxima segunda-feira 00:00
	// ---------------------------------------------------------

	diasDesdeSegunda := (int(inicioHoje.Weekday()) + 6) % 7

	inicioSemana := inicioHoje.AddDate(
		0,
		0,
		-diasDesdeSegunda,
	)

	fimSemana := inicioSemana.AddDate(0, 0, 7)

	// ---------------------------------------------------------
	// MÊS ATUAL
	// ---------------------------------------------------------

	inicioMesAtual := time.Date(
		agora.Year(),
		agora.Month(),
		1,
		0,
		0,
		0,
		0,
		time.Local,
	)

	fimMesAtual := inicioMesAtual.AddDate(0, 1, 0)

	// ---------------------------------------------------------
	// MÊS FILTRADO
	// ---------------------------------------------------------

	inicioMesFiltro := time.Date(
		mesFiltro.Year(),
		mesFiltro.Month(),
		1,
		0,
		0,
		0,
		0,
		time.Local,
	)

	fimMesFiltro := inicioMesFiltro.AddDate(0, 1, 0)

	// ---------------------------------------------------------
	// CONSULTAS
	// ---------------------------------------------------------

	hoje, err := s.calcularPeriodo(
		ctx,
		req,
		profissional,
		inicioHoje,
		fimHoje,
		false,
	)
	if err != nil {
		return ResponseData{}, err
	}

	semana, err := s.calcularPeriodo(
		ctx,
		req,
		profissional,
		inicioSemana,
		fimSemana,
		false,
	)
	if err != nil {
		return ResponseData{}, err
	}

	mesAtual, err := s.calcularPeriodo(
		ctx,
		req,
		profissional,
		inicioMesAtual,
		fimMesAtual,
		true,
	)
	if err != nil {
		return ResponseData{}, err
	}

	mesFiltroPeriodo, err := s.calcularPeriodo(
		ctx,
		req,
		profissional,
		inicioMesFiltro,
		fimMesFiltro,
		true,
	)
	if err != nil {
		return ResponseData{}, err
	}

	return ResponseData{
		Nome:                profissional.Nome,
		Especialidade:       profissional.Especialidade,
		FotoURL:             profissional.FotoURL,
		TipoRemuneracao:     profissional.TipoRemuneracao,
		PercentualComissao:  profissional.PercentualComissao,
		SalarioFixo:         profissional.SalarioFixo,
		FrequenciaPagamento: profissional.FrequenciaPagamento,

		Hoje:      hoje,
		Semana:    semana,
		MesAtual:  mesAtual,
		MesFiltro: mesFiltroPeriodo,
	}, nil
}

func (s *Service) calcularPeriodo(
	ctx context.Context,
	req Request,
	profissional ProfessionalSalary,
	inicio time.Time,
	fim time.Time,
	aplicaSalarioFixo bool,
) (Periodo, error) {

	dados, err := s.repository.BuscarPeriodo(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
		inicio,
		fim,
	)

	if err != nil {
		return Periodo{}, fmt.Errorf(
			"buscar período financeiro: %w",
			err,
		)
	}

	periodo := Periodo{
		Qtd:         dados.Qtd,
		TotalGerado: arredondar(dados.TotalGerado),
	}

	switch profissional.TipoRemuneracao {

	case "comissao":

		periodo.Comissao = calcularComissao(
			dados.TotalGerado,
			profissional.PercentualComissao,
		)

		periodo.ValorGanho = periodo.Comissao

	case "fixo":

		// O salário fixo não deve ser repetido em
		// "Hoje" e "Semana" como se fosse um pagamento.
		//
		// Para o mês, mostramos o salário contratado.
		if aplicaSalarioFixo {
			periodo.SalarioFixo = arredondar(
				profissional.SalarioFixo,
			)

			periodo.ValorGanho = periodo.SalarioFixo
		}

	case "misto":

		periodo.Comissao = calcularComissao(
			dados.TotalGerado,
			profissional.PercentualComissao,
		)

		if aplicaSalarioFixo {
			periodo.SalarioFixo = arredondar(
				profissional.SalarioFixo,
			)

			periodo.ValorGanho = arredondar(
				periodo.SalarioFixo + periodo.Comissao,
			)
		} else {
			// Fora do mês, mostramos somente a
			// parte variável gerada pelo período.
			periodo.ValorGanho = periodo.Comissao
		}

	default:
		return Periodo{}, fmt.Errorf(
			"tipo de remuneração inválido: %s",
			profissional.TipoRemuneracao,
		)
	}

	return periodo, nil
}

func calcularComissao(
	faturamento float64,
	percentual float64,
) float64 {

	if faturamento <= 0 || percentual <= 0 {
		return 0
	}

	return arredondar(
		faturamento * percentual / 100,
	)
}

func arredondar(valor float64) float64 {
	return math.Round(valor*100) / 100
}