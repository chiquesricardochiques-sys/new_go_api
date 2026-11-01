package verAgendamentosAvulso

import (
    "context"
    "errors"
    "strings"
)

type Service struct {
    repository *Repository
}

func NewService(repository *Repository) *Service {
    return &Service{
        repository: repository,
    }
}

func (s *Service) Listar(
    ctx context.Context,
    filtro FiltroAgendamentosAvulsos,
) ([]AgendamentoAvulso, error) {

    if filtro.SalaoID <= 0 {
        return nil, errors.New("salão inválido")
    }

    filtro.Data = strings.TrimSpace(filtro.Data)

    if filtro.Data == "" {
        return nil, errors.New("informe a data")
    }

    return s.repository.Listar(ctx, filtro)
}

func (s *Service) Cancelar(
    ctx context.Context,
    req CancelarAgendamentoRequest,
) error {

    if req.SalaoID <= 0 {
        return errors.New("salão inválido")
    }

    if req.AgendamentoID <= 0 {
        return errors.New("agendamento inválido")
    }

    return s.repository.Cancelar(
        ctx,
        req.SalaoID,
        req.AgendamentoID,
    )
}