package statusAG

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "time/tzdata"
	"github.com/ricardo-platform/salao-usuarios-api/internal/modules/profissional/profissional"
)

var fusoBrasil = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*60*60)
	}
	return loc
}()

type Service struct {
	repository *Repository
}

func NewService(repository *Repository, db *sql.DB) *Service {
	return &Service{
		repository: repository,
	}
}

func agora() time.Time {
	return time.Now().In(fusoBrasil)
}

func validarBase(b BaseRequest) error {
	if b.IDSalao <= 0 {
		return fmt.Errorf("%w: id_salao obrigatório", profissional.ErrDadosInvalidos)
	}
	if b.ProfissionalID <= 0 {
		return fmt.Errorf("%w: profissional_id obrigatório", profissional.ErrDadosInvalidos)
	}
	return nil
}

func (s *Service) IniciarCorte(ctx context.Context, req IniciarCorteRequest) (StatusResponse, error) {
	if err := validarBase(req.BaseRequest); err != nil {
		return StatusResponse{}, err
	}
	if req.AgendamentoID <= 0 {
		return StatusResponse{}, fmt.Errorf("%w: agendamento_id obrigatório", profissional.ErrDadosInvalidos)
	}

	if _, err := s.repository.StatusDoAgendamento(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID); err != nil {
		return StatusResponse{}, err
	}

	agoraReal := agora()
	if err := s.repository.IniciarCorteRepo(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID, agoraReal); err != nil {
		return StatusResponse{}, err
	}

	return StatusResponse{NovoStatus: "confirmado"}, nil
}

func (s *Service) FinalizarCorte(ctx context.Context, req FinalizarCorteRequest) (StatusResponse, error) {
	if err := validarBase(req.BaseRequest); err != nil {
		return StatusResponse{}, err
	}
	if req.AgendamentoID <= 0 {
		return StatusResponse{}, fmt.Errorf("%w: agendamento_id obrigatório", profissional.ErrDadosInvalidos)
	}

	if _, err := s.repository.StatusDoAgendamento(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID); err != nil {
		return StatusResponse{}, err
	}

	agoraReal := agora()
	if err := s.repository.FinalizarCorteRepo(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID, agoraReal); err != nil {
		return StatusResponse{}, err
	}

	return StatusResponse{NovoStatus: "concluido"}, nil
}

func (s *Service) CancelarAgendamento(ctx context.Context, req CancelarAgendamentoRequest) (StatusResponse, error) {
	if err := validarBase(req.BaseRequest); err != nil {
		return StatusResponse{}, err
	}
	if req.AgendamentoID <= 0 {
		return StatusResponse{}, fmt.Errorf("%w: agendamento_id obrigatório", profissional.ErrDadosInvalidos)
	}

	if _, err := s.repository.StatusDoAgendamento(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID); err != nil {
		return StatusResponse{}, err
	}

	if err := s.repository.CancelarAgendamentoRepo(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID, req.Motivo); err != nil {
		return StatusResponse{}, err
	}

	return StatusResponse{NovoStatus: "cancelado"}, nil
}

func (s *Service) NaoCompareceu(ctx context.Context, req NaoCompareceuRequest) (StatusResponse, error) {
	if err := validarBase(req.BaseRequest); err != nil {
		return StatusResponse{}, err
	}
	if req.AgendamentoID <= 0 {
		return StatusResponse{}, fmt.Errorf("%w: agendamento_id obrigatório", profissional.ErrDadosInvalidos)
	}

	if _, err := s.repository.StatusDoAgendamento(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID); err != nil {
		return StatusResponse{}, err
	}

	if err := s.repository.NaoCompareceuRepo(ctx, req.IDSalao, req.ProfissionalID, req.AgendamentoID); err != nil {
		return StatusResponse{}, err
	}

	return StatusResponse{NovoStatus: "nao_compareceu"}, nil
}