package saloes

import (
	"context"
	"regexp"
	"strings"
)

const (
	limitePadrao = 20
	limiteMaximo = 50
)

var regexSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

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
	slug := strings.ToLower(strings.TrimSpace(req.Slug))

	if slug == "" || len(slug) > 180 || !regexSlug.MatchString(slug) {
		return nil, ErrDadosInvalidos
	}

	return s.repository.BuscarPorSlug(ctx, slug)
}

func (s *Service) Cidades(ctx context.Context) ([]Cidade, error) {
	return s.repository.ListarCidades(ctx)
}