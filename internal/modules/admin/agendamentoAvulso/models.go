package agendamentoavulso

import "time"

type SchedulePageRequest struct {
	SalaoID        int64  `json:"salao_id"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
}

type HorariosDisponiveisRequest struct {
	SalaoID        int64  `json:"salao_id"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
}

type ServicosProfissionalRequest struct {
	SalaoID        int64 `json:"salao_id"`
	ProfissionalID int64 `json:"profissional_id"`
}

type AgendarRequest struct {
	SalaoID         int64    `json:"salao_id"`
	NomeCliente     string   `json:"nome_cliente"`
	TelefoneCliente string   `json:"telefone_cliente"`
	Descricao       *string  `json:"descricao"`
	ServicoIDs      []int64  `json:"servico_ids"`
	ProfissionalID  int64    `json:"profissional_id"`
	Data            string   `json:"data"`
	Hora            string   `json:"hora"`
}

type Horario struct {
	Hora          string   `json:"hora"`
	Disponivel    bool     `json:"disponivel"`
	AgendamentoID *int64   `json:"agendamento,omitempty"`
	Servicos      []string `json:"servicos,omitempty"`
}

type Servico struct {
	ID             int64   `json:"id"`
	Nome           string  `json:"nome"`
	Descricao      *string `json:"descricao"`
	FotoURL        *string `json:"foto_url"`
	Preco          float64 `json:"preco"`
	DuracaoMinutos int     `json:"duracao_minutos"`
}

type Profissional struct {
	ID            int64   `json:"id"`
	Nome          string  `json:"nome"`
	Especialidade *string `json:"especialidade"`
	FotoURL       *string `json:"foto_url"`
	UsaHorarios   bool    `json:"usa_horarios"`
}

type ScheduleData struct {
	Textos               map[string]string `json:"textos"`
	Servicos             []Servico         `json:"servicos"`
	Profissionais        []Profissional    `json:"profissionais"`
	CurrentDate          string            `json:"currentDate"`
	UsaHorarios          int               `json:"usa_horarios"`
	HorariosDisponiveis  []Horario         `json:"horariosDisponiveis"`
	SelectedProfissional *int64            `json:"selectedProfissional"`
	SelectedData         string            `json:"selectedData"`
}

type SchedulePageResponse struct {
	Success bool         `json:"success"`
	Data    *ScheduleData `json:"data,omitempty"`
}

type HorariosDisponiveisResponse struct {
	Success bool `json:"success"`
	Data    struct {
		UsaHorarios int       `json:"usa_horarios"`
		Horarios    []Horario `json:"horarios"`
	} `json:"data"`
}

type AgendarResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *struct {
		AgendamentoID int64 `json:"agendamento_id"`
	} `json:"data,omitempty"`
}

type Grade struct {
	Abre        bool
	Abertura    string
	PausaInicio *string
	PausaFim    *string
	Fechamento  string
}

type AgendamentoExistente struct {
	ID       int64
	Inicio   time.Time
	Fim      time.Time
	Servicos []string
}

type BloqueioExistente struct {
	ID     int64
	Inicio time.Time
	Fim    time.Time
}