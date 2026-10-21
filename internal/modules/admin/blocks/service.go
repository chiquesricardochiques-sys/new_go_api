package blocks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDadosInvalidos = errors.New("dados inválidos")
	ErrDataPassada    = errors.New("não é possível bloquear datas passadas")
)

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

func (s *Service) hoje() time.Time {
	agora := time.Now().In(s.location)
	return time.Date(
		agora.Year(), agora.Month(), agora.Day(),
		0, 0, 0, 0, s.location,
	)
}

func validarSalaoProfissional(idSalao, profissionalID int64) error {
	if idSalao <= 0 || profissionalID <= 0 {
		return fmt.Errorf(
			"%w: id_salao e profissional_id são obrigatórios",
			ErrDadosInvalidos,
		)
	}
	return nil
}

func validarData(data string) (time.Time, error) {
	data = strings.TrimSpace(data)

	resultado, err := time.Parse("2006-01-02", data)
	if err != nil || resultado.Format("2006-01-02") != data {
		return time.Time{}, fmt.Errorf(
			"%w: data deve estar no formato AAAA-MM-DD",
			ErrDadosInvalidos,
		)
	}

	return resultado, nil
}

func (s *Service) verificarProfissional(
	ctx context.Context,
	idSalao, profissionalID int64,
) error {
	existe, err := s.repository.ProfissionalExiste(
		ctx, idSalao, profissionalID,
	)
	if err != nil {
		return err
	}
	if !existe {
		return ErrProfissionalNaoEncontrado
	}
	return nil
}

func (s *Service) Listar(
	ctx context.Context,
	req ListRequest,
) (*ListResponse, error) {
	if err := validarSalaoProfissional(
		req.IDSalao, req.ProfissionalID,
	); err != nil {
		return nil, err
	}

	if err := s.verificarProfissional(
		ctx, req.IDSalao, req.ProfissionalID,
	); err != nil {
		return nil, err
	}

	hoje := s.hoje()
	dataSelecionada := hoje.Format("2006-01-02")

	if strings.TrimSpace(req.Data) != "" {
		data, err := validarData(req.Data)
		if err != nil {
			return nil, err
		}
		dataSelecionada = data.Format("2006-01-02")
	}

	dias, err := s.repository.ListarDias(
		ctx, req.IDSalao, req.ProfissionalID, hoje,
	)
	if err != nil {
		return nil, err
	}

	horas, err := s.repository.ListarHoras(
		ctx, req.IDSalao, req.ProfissionalID, dataSelecionada,
	)
	if err != nil {
		return nil, err
	}

	return &ListResponse{
		DiasBloqueados:  dias,
		HorasBloqueadas: horas,
		Data:            dataSelecionada,
	}, nil
}

func (s *Service) AdicionarDia(
	ctx context.Context,
	req AddDayRequest,
) error {
	if err := validarSalaoProfissional(
		req.IDSalao, req.ProfissionalID,
	); err != nil {
		return err
	}

	data, err := validarData(req.Data)
	if err != nil {
		return err
	}

	if data.Before(s.hoje()) {
		return ErrDataPassada
	}

	if err := s.verificarProfissional(
		ctx, req.IDSalao, req.ProfissionalID,
	); err != nil {
		return err
	}

	return s.repository.AdicionarDia(
		ctx,
		req.IDSalao,
		req.ProfissionalID,
		data.Format("2006-01-02"),
		strings.TrimSpace(req.Motivo),
	)
}

func (s *Service) AdicionarHora(
	ctx context.Context,
	req AddHourRequest,
) error {
	if err := validarSalaoProfissional(
		req.IDSalao, req.ProfissionalID,
	); err != nil {
		return err
	}

	data, err := validarData(req.Data)
	if err != nil {
		return err
	}

	if data.Before(s.hoje()) {
		return ErrDataPassada
	}

	hora := strings.TrimSpace(req.Hora)
	if _, err := time.Parse("15:04", hora); err != nil {
		if _, err := time.Parse("15:04:05", hora); err != nil {
			return fmt.Errorf(
				"%w: hora deve estar no formato HH:MM ou HH:MM:SS",
				ErrDadosInvalidos,
			)
		}
	}

	if err := s.verificarProfissional(
		ctx, req.IDSalao, req.ProfissionalID,
	); err != nil {
		return err
	}

	if len(hora) == 5 {
		hora += ":00"
	}

	return s.repository.AdicionarHora(
		ctx,
		req.IDSalao,
		req.ProfissionalID,
		data.Format("2006-01-02"),
		hora,
		strings.TrimSpace(req.Motivo),
	)
}

func (s *Service) RemoverDia(
	ctx context.Context,
	idSalao, id int64,
) error {
	if idSalao <= 0 || id <= 0 {
		return fmt.Errorf("%w: id_salao e id são obrigatórios", ErrDadosInvalidos)
	}
	return s.repository.RemoverDia(ctx, idSalao, id)
}

func (s *Service) RemoverHora(
	ctx context.Context,
	idSalao, id int64,
) error {
	if idSalao <= 0 || id <= 0 {
		return fmt.Errorf("%w: id_salao e id são obrigatórios", ErrDadosInvalidos)
	}
	return s.repository.RemoverHora(ctx, idSalao, id)
}
