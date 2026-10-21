package services

import (
	"context"
	"database/sql"
	"fmt"
)

type RepositoryInterface interface {
	ListGlobalServices(ctx context.Context) ([]Servico, error)

	ListSalonServices(
		ctx context.Context,
		salaoID uint64,
	) ([]ServicoSalao, error)

	GetSalonService(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
	) (*ServicoSalao, error)

	AddSalonService(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
		preco float64,
		duracaoMinutos uint,
	) error

	DeactivateSalonService(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
	) error

	UnlinkSalonService(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
	) error

	ListProfessionals(
		ctx context.Context,
		salaoID uint64,
	) ([]Profissional, error)

	ListProfessionalsByService(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
	) ([]Profissional, error)

	ListProfessionalServices(
		ctx context.Context,
		salaoID uint64,
		profissionalID uint64,
	) ([]ServicoProfissionalItem, error)

	ListLinkedProfessionalServices(
		ctx context.Context,
		salaoID uint64,
		profissionalID uint64,
	) ([]ServicoProfissionalItem, error)

	AddProfessionalService(
		ctx context.Context,
		salaoID uint64,
		profissionalID uint64,
		servicoID uint64,
	) error

	RemoveProfessionalService(
		ctx context.Context,
		salaoID uint64,
		profissionalID uint64,
		servicoID uint64,
	) error

	ValidateGlobalService(
		ctx context.Context,
		servicoID uint64,
	) error

	ValidateSalonService(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
	) error

	ValidateSalonServiceRelation(
		ctx context.Context,
		salaoID uint64,
		servicoID uint64,
	) error

	ValidateProfessional(
		ctx context.Context,
		salaoID uint64,
		profissionalID uint64,
	) error
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{
		repo: repo,
	}
}

// ============================================================
// SERVIÇOS GLOBAIS
// ============================================================

func (s *Service) ListGlobalServices(
	ctx context.Context,
) ([]Servico, error) {

	return s.repo.ListGlobalServices(ctx)
}

// ============================================================
// SERVIÇOS DO SALÃO
// ============================================================

func (s *Service) ListSalonServices(
	ctx context.Context,
	salaoID uint64,
) ([]ServicoSalao, error) {

	if salaoID == 0 {
		return nil, fmt.Errorf("salao_id é obrigatório")
	}

	return s.repo.ListSalonServices(ctx, salaoID)
}

func (s *Service) GetSalonService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) (*ServicoSalao, error) {

	if salaoID == 0 {
		return nil, fmt.Errorf("salao_id é obrigatório")
	}

	if servicoID == 0 {
		return nil, fmt.Errorf("servico_id é obrigatório")
	}

	return s.repo.GetSalonService(
		ctx,
		salaoID,
		servicoID,
	)
}

func (s *Service) AddSalonService(
	ctx context.Context,
	req AddSalonServiceRequest,
) error {

	if req.SalaoID == 0 {
		return fmt.Errorf("salao_id é obrigatório")
	}

	if req.ServicoID == 0 {
		return fmt.Errorf("servico_id é obrigatório")
	}

	if req.Preco < 0 {
		return fmt.Errorf("preço não pode ser negativo")
	}

	if req.DuracaoMinutos == 0 {
		return fmt.Errorf("duração deve ser maior que zero")
	}

	if err := s.repo.ValidateGlobalService(
		ctx,
		req.ServicoID,
	); err != nil {
		return err
	}

	return s.repo.AddSalonService(
		ctx,
		req.SalaoID,
		req.ServicoID,
		req.Preco,
		req.DuracaoMinutos,
	)
}

func (s *Service) DeactivateSalonService(
	ctx context.Context,
	req DeactivateSalonServiceRequest,
) error {

	if req.SalaoID == 0 {
		return fmt.Errorf("salao_id é obrigatório")
	}

	if req.ServicoID == 0 {
		return fmt.Errorf("servico_id é obrigatório")
	}

	if err := s.repo.ValidateSalonServiceRelation(
		ctx,
		req.SalaoID,
		req.ServicoID,
	); err != nil {
		return err
	}

	return s.repo.DeactivateSalonService(
		ctx,
		req.SalaoID,
		req.ServicoID,
	)
}

func (s *Service) UnlinkSalonService(
	ctx context.Context,
	req UnlinkSalonServiceRequest,
) error {

	if req.SalaoID == 0 {
		return fmt.Errorf("salao_id é obrigatório")
	}

	if req.ServicoID == 0 {
		return fmt.Errorf("servico_id é obrigatório")
	}

	if err := s.repo.ValidateSalonServiceRelation(
		ctx,
		req.SalaoID,
		req.ServicoID,
	); err != nil {
		return err
	}

	return s.repo.UnlinkSalonService(
		ctx,
		req.SalaoID,
		req.ServicoID,
	)
}

// ============================================================
// PROFISSIONAIS
// ============================================================

func (s *Service) ListProfessionals(
	ctx context.Context,
	salaoID uint64,
) ([]Profissional, error) {

	if salaoID == 0 {
		return nil, fmt.Errorf("salao_id é obrigatório")
	}

	return s.repo.ListProfessionals(ctx, salaoID)
}

func (s *Service) ListProfessionalsByService(
	ctx context.Context,
	salaoID uint64,
	servicoID uint64,
) ([]Profissional, error) {

	if salaoID == 0 {
		return nil, fmt.Errorf("salao_id é obrigatório")
	}

	if servicoID == 0 {
		return nil, fmt.Errorf("servico_id é obrigatório")
	}

	if err := s.repo.ValidateSalonService(
		ctx,
		salaoID,
		servicoID,
	); err != nil {
		return nil, err
	}

	return s.repo.ListProfessionalsByService(
		ctx,
		salaoID,
		servicoID,
	)
}

// ============================================================
// SERVIÇOS DO PROFISSIONAL
// ============================================================

func (s *Service) ListProfessionalServices(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
) ([]ServicoProfissionalItem, error) {

	if salaoID == 0 {
		return nil, fmt.Errorf("salao_id é obrigatório")
	}

	if profissionalID == 0 {
		return nil, fmt.Errorf("profissional_id é obrigatório")
	}

	if err := s.repo.ValidateProfessional(
		ctx,
		salaoID,
		profissionalID,
	); err != nil {
		return nil, err
	}

	return s.repo.ListProfessionalServices(
		ctx,
		salaoID,
		profissionalID,
	)
}

func (s *Service) ListLinkedProfessionalServices(
	ctx context.Context,
	salaoID uint64,
	profissionalID uint64,
) ([]ServicoProfissionalItem, error) {

	if salaoID == 0 {
		return nil, fmt.Errorf("salao_id é obrigatório")
	}

	if profissionalID == 0 {
		return nil, fmt.Errorf("profissional_id é obrigatório")
	}

	if err := s.repo.ValidateProfessional(
		ctx,
		salaoID,
		profissionalID,
	); err != nil {
		return nil, err
	}

	return s.repo.ListLinkedProfessionalServices(
		ctx,
		salaoID,
		profissionalID,
	)
}

func (s *Service) AddProfessionalService(
	ctx context.Context,
	req AddProfessionalServiceRequest,
) error {

	if req.SalaoID == 0 {
		return fmt.Errorf("salao_id é obrigatório")
	}

	if req.ProfissionalID == 0 {
		return fmt.Errorf("profissional_id é obrigatório")
	}

	if req.ServicoID == 0 {
		return fmt.Errorf("servico_id é obrigatório")
	}

	if err := s.repo.ValidateProfessional(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
	); err != nil {
		return err
	}

	if err := s.repo.ValidateSalonService(
		ctx,
		req.SalaoID,
		req.ServicoID,
	); err != nil {
		return err
	}

	return s.repo.AddProfessionalService(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
		req.ServicoID,
	)
}

func (s *Service) RemoveProfessionalService(
	ctx context.Context,
	req RemoveProfessionalServiceRequest,
) error {

	if req.SalaoID == 0 {
		return fmt.Errorf("salao_id é obrigatório")
	}

	if req.ProfissionalID == 0 {
		return fmt.Errorf("profissional_id é obrigatório")
	}

	if req.ServicoID == 0 {
		return fmt.Errorf("servico_id é obrigatório")
	}

	if err := s.repo.ValidateProfessional(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
	); err != nil {
		return err
	}

	return s.repo.RemoveProfessionalService(
		ctx,
		req.SalaoID,
		req.ProfissionalID,
		req.ServicoID,
	)
}

// ============================================================
// AUXILIAR
// ============================================================

func isNotFound(err error) bool {
	return err == sql.ErrNoRows
}