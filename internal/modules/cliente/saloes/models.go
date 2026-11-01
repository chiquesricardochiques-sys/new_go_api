package saloes

import "errors"

var (
	ErrDadosInvalidos     = errors.New("dados inválidos")
	ErrSalaoNaoEncontrado = errors.New("salão não encontrado")
)

type BuscarRequest struct {
	Busca      string   `json:"busca"`
	Cidade     string   `json:"cidade"`
	Estado     string   `json:"estado"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	AuthUserID int64    `json:"auth_user_id"` // opcional: marca salões onde o cliente tem horário
	Limite     int      `json:"limite"`
	Offset     int      `json:"offset"`
}

type PorSlugRequest struct {
	Slug string `json:"slug"`
}

type SalaoItem struct {
	ID             int64    `json:"id"`
	Nome           string   `json:"nome"`
	Slug           string   `json:"slug"`
	Endereco       *string  `json:"endereco"`
	Cidade         *string  `json:"cidade"`
	Estado         *string  `json:"estado"`
	Whatsapp       *string  `json:"whatsapp"`
	LogoURL        *string  `json:"logo_url"`
	DistanciaKm    *float64 `json:"distancia_km"`
	TemAgendamento bool     `json:"tem_agendamento"`
}

type BuscarData struct {
	Saloes  []SalaoItem `json:"saloes"`
	Limite  int         `json:"limite"`
	Offset  int         `json:"offset"`
	TemMais bool        `json:"tem_mais"`
}

type Cidade struct {
	Cidade string `json:"cidade"`
	Estado string `json:"estado"`
	Total  int    `json:"total"`
}

// filtro interno (já validado pelo service)
type filtro struct {
	Busca      string
	Cidade     string
	Estado     string
	Latitude   *float64
	Longitude  *float64
	AuthUserID int64
	Limite     int
	Offset     int
}