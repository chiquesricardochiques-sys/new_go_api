package saloes

import (
	"context"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata" // garante o fuso em containers sem tzdata
)

const (
	limitePadrao = 20
	limiteMaximo = 50
)

var regexSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Mesmo valor do enum horario_funcionamento.dia_semana, indexado por time.Weekday
var diasSemana = [...]string{"dom", "seg", "ter", "qua", "qui", "sex", "sab"}

var fusoBrasil = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*3600)
	}
	return loc
}()

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func truncar(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > max {
		r = r[:max]
	}
	return string(r)
}

func slugValido(s string) (string, bool) {
	slug := strings.ToLower(strings.TrimSpace(s))
	ok := slug != "" && len(slug) <= 180 && regexSlug.MatchString(slug)
	return slug, ok
}

func (s *Service) Buscar(
	ctx context.Context,
	req BuscarRequest,
) (*BuscarData, error) {
	// localização: ou vêm as duas coordenadas ou nenhuma
	if (req.Latitude == nil) != (req.Longitude == nil) {
		return nil, ErrDadosInvalidos
	}

	if req.Latitude != nil {
		if *req.Latitude < -90 || *req.Latitude > 90 ||
			*req.Longitude < -180 || *req.Longitude > 180 {
			return nil, ErrDadosInvalidos
		}
	}

	estado := strings.ToUpper(strings.TrimSpace(req.Estado))
	if estado != "" && len(estado) != 2 {
		return nil, ErrDadosInvalidos
	}

	limite := req.Limite
	if limite <= 0 {
		limite = limitePadrao
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	agora := time.Now().In(fusoBrasil)

	// pede 1 a mais só para saber se existe próxima página
	itens, err := s.repository.Buscar(ctx, filtro{
		Busca:      truncar(req.Busca, 100),
		Cidade:     truncar(req.Cidade, 100),
		Estado:     estado,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		AuthUserID: req.AuthUserID,
		Limite:     limite + 1,
		Offset:     offset,
		Dia:        diasSemana[agora.Weekday()],
		Hora:       agora.Format("15:04:05"),
	})
	if err != nil {
		return nil, err
	}

	temMais := len(itens) > limite
	if temMais {
		itens = itens[:limite]
	}

	return &BuscarData{
		Saloes:  itens,
		Limite:  limite,
		Offset:  offset,
		TemMais: temMais,
	}, nil
}

func (s *Service) PorSlug(
	ctx context.Context,
	req PorSlugRequest,
) (*SalaoItem, error) {
	slug, ok := slugValido(req.Slug)
	if !ok {
		return nil, ErrDadosInvalidos
	}

	return s.repository.BuscarPorSlug(ctx, slug)
}

func (s *Service) Detalhes(
	ctx context.Context,
	req PorSlugRequest,
) (*DetalhesData, error) {
	slug, ok := slugValido(req.Slug)
	if !ok {
		return nil, ErrDadosInvalidos
	}

	return s.repository.Detalhes(ctx, slug)
}

func (s *Service) Avaliar(
	ctx context.Context,
	req AvaliarRequest,
) error {
	if req.AuthUserID <= 0 || req.AgendamentoID <= 0 ||
		req.Nota < 1 || req.Nota > 5 {
		return ErrDadosInvalidos
	}

	req.Comentario = truncar(req.Comentario, 500)

	return s.repository.Avaliar(ctx, req)
}

func (s *Service) Cidades(ctx context.Context) ([]Cidade, error) {
	return s.repository.ListarCidades(ctx)
}