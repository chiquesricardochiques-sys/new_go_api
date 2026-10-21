package salary

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidData = errors.New("dados inválidos")

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Rendimento(ctx context.Context, req Request) (ResponseData, error) {
	if req.SalaoID <= 0 || req.ProfissionalID <= 0 {
		return ResponseData{}, ErrInvalidData
	}

	mesFiltro, err := time.ParseInLocation("2006-01", req.MesAno, time.Local)
	if err != nil || mesFiltro.Year() < 2000 || mesFiltro.Year() > 2200 {
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
	inicioHoje := time.Date(
		agora.Year(), agora.Month(), agora.Day(),
		0, 0, 0, 0, time.Local,
	)
	inicioAmanha := inicioHoje.AddDate(0, 0, 1)

	inicioSemana := inicioHoje.AddDate(0, 0, -((int(inicioHoje.Weekday()) + 6) % 7))
	fimSemana := inicioSemana.AddDate(0, 0, 7)

	inicioMesAtual := time.Date(agora.Year(), agora.Month(), 1, 0, 0, 0, 0, time.Local)
	fimMesAtual := inicioMesAtual.AddDate(0, 1, 0)

	inicioMesFiltro := time.Date(mesFiltro.Year(), mesFiltro.Month(), 1, 0, 0, 0, 0, time.Local)
	fimMesFiltro := inicioMesFiltro.AddDate(0, 1, 0)

	hoje, err := s.calcularPeriodo(ctx, req, profissional, inicioHoje, inicioAmanha)
	if err != nil {
		return ResponseData{}, err
	}

	semana, err := s.calcularPeriodo(ctx, req, profissional, inicioSemana, fimSemana)
	if err != nil {
		return ResponseData{}, err
	}

	mesAtual, err := s.calcularPeriodo(ctx, req, profissional, inicioMesAtual, fimMesAtual)
	if err != nil {
		return ResponseData{}, err
	}

	mesSelecionado, err := s.calcularPeriodo(ctx, req, profissional, inicioMesFiltro, fimMesFiltro)
	if err != nil {
		return ResponseData{}, err
	}

	return ResponseData{
		Nome:               profissional.Nome,
		FotoURL:            profissional.FotoURL,
		TipoRemuneracao:    profissional.TipoRemuneracao,
		PercentualComissao: profissional.PercentualComissao,
		SalarioFixo:        profissional.SalarioFixo,
		Hoje:               hoje,
		Semana:             semana,
		MesAtual:           mesAtual,
		MesFiltro:          mesSelecionado,
	}, nil
}

func (s *Service) calcularPeriodo(
	ctx context.Context,
	req Request,
	profissional ProfessionalSalary,
	inicio, fim time.Time,
) (Periodo, error) {
	agendamentos, err := s.repository.BuscarRendimentos(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
		inicio,
		fim,
	)
	if err != nil {
		return Periodo{}, fmt.Errorf("buscar rendimentos do período: %w", err)
	}

	var total float64
	for _, agendamento := range agendamentos {
		total += agendamento.Total
	}

	periodo := Periodo{
		Qtd:         len(agendamentos),
		TotalGerado: total,
	}

	// A comissão é calculada apenas para remuneração por comissão.
	// As regras de pagamento fixo e misto precisam ser confirmadas.
	if profissional.TipoRemuneracao == "comissao" {
		periodo.ValorGanho = total * profissional.PercentualComissao / 100
	}

	return periodo, nil
}
