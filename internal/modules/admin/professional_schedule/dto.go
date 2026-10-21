package professional_schedule

type ScheduleDay struct {
	DiaSemana   string `json:"dia_semana"`
	Abre        bool   `json:"abre"`
	Abertura    string `json:"abertura"`
	PausaInicio string `json:"pausa_inicio"`
	PausaFim    string `json:"pausa_fim"`
	Fechamento  string `json:"fechamento"`
}

type ProfessionalSchedule struct {
	ProfissionalID uint64        `json:"profissional_id"`
	Grade          []ScheduleDay `json:"grade"`
}

type GetScheduleRequest struct {
	SalaoID        uint64 `json:"salao_id"`
	ProfissionalID uint64 `json:"profissional_id"`
}

type UpdateScheduleRequest struct {
	SalaoID        uint64        `json:"salao_id"`
	ProfissionalID uint64        `json:"profissional_id"`
	Grade          []ScheduleDay `json:"grade"`
}

type ResetScheduleRequest struct {
	SalaoID        uint64 `json:"salao_id"`
	ProfissionalID uint64 `json:"profissional_id"`
}