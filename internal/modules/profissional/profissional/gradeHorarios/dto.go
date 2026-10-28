package gradehorarios

import "errors"

// Mensagens em português porque o app mostra err.message direto pro profissional.
var (
	ErrDadosInvalidos          = errors.New("dados inválidos")
	ErrNaoEncontrado           = errors.New("registro não encontrado")
	ErrProfissionalSemHorarios = errors.New("este profissional não trabalha com grade de horários")
	ErrHorarioPassado          = errors.New("não é possível bloquear um horário que já passou")
	ErrForaDaGrade             = errors.New("esse horário está fora da grade do profissional")
	ErrDiaBloqueado            = errors.New("este dia já está bloqueado")
	ErrTemAgendamento          = errors.New("já existe agendamento nesse horário; só é possível bloquear horários livres")
	ErrJaBloqueado             = errors.New("esse horário já está bloqueado")
)

// ---------- requests ----------

type BaseRequest struct {
	IDSalao        int64 `json:"id_salao"`
	ProfissionalID int64 `json:"profissional_id"`
}

type HorariosDiaRequest struct {
	BaseRequest
	Data string `json:"data"` // YYYY-MM-DD
}

type BloquearRequest struct {
	BaseRequest
	Data       string `json:"data"`        // YYYY-MM-DD
	HoraInicio string `json:"hora_inicio"` // HH:MM
	HoraFim    string `json:"hora_fim"`    // HH:MM (exclusivo)
	Motivo     string `json:"motivo"`      // opcional
}

type DesbloquearRequest struct {
	BaseRequest
	BloqueioID int64 `json:"bloqueio_id"`
}

// ---------- responses ----------

const (
	StatusLivre     = "livre"
	StatusAgendado  = "agendado"
	StatusBloqueado = "bloqueado"
)

// Slot de 30 minutos da grade do profissional.
//
//	livre:     pode bloquear
//	agendado:  tem cliente marcado (não pode bloquear)
//	bloqueado: o próprio profissional bloqueou (pode desbloquear)
type Slot struct {
	Hora       string `json:"hora"`
	Status     string `json:"status"`
	Disponivel bool   `json:"disponivel"` // = (status == livre), compatível com a tela antiga

	// quando agendado
	Agendamento *int64   `json:"agendamento"`
	Servicos    []string `json:"servicos"`

	// quando bloqueado
	BloqueioID     *int64  `json:"bloqueio_id"`
	BloqueioInicio *string `json:"bloqueio_inicio"` // HH:MM
	BloqueioFim    *string `json:"bloqueio_fim"`    // HH:MM
	Motivo         *string `json:"motivo"`
}

type HorariosDiaResponse struct {
	UsaHorarios int    `json:"usa_horarios"`
	Bloqueado   bool   `json:"bloqueado"` // dia inteiro bloqueado
	Horarios    []Slot `json:"horarios"`
}

type BloquearResponse struct {
	BloqueioID int64  `json:"bloqueio_id"`
	Data       string `json:"data"`
	HoraInicio string `json:"hora_inicio"`
	HoraFim    string `json:"hora_fim"`
}

type DesbloquearResponse struct {
	BloqueioID int64 `json:"bloqueio_id"`
}

// ---------- tipos internos (repository <-> service) ----------

type Profissional struct {
	ID          int64
	UsaHorarios int
}

type Grade struct {
	Abre        bool
	Abertura    string // HH:MM
	PausaInicio string // HH:MM ou ""
	PausaFim    string // HH:MM ou ""
	Fechamento  string // HH:MM
}

// AgendamentoDia: quem ocupa o horário (hora_marcada + duracao_minutos).
type AgendamentoDia struct {
	ID       int64
	Inicio   string // "YYYY-MM-DD HH:MM:SS"
	Duracao  int
	Servicos []string
}

type BloqueioDia struct {
	ID     int64
	Inicio string // "YYYY-MM-DD HH:MM:SS"
	Fim    string
	Motivo *string
}