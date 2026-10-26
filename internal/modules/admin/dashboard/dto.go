package dashboard

type ResumoResponse struct {
	Agendamentos    AgendamentosResumo      `json:"agendamentos"`
	Faturamento     FaturamentoResumo       `json:"faturamento"`
	FaturamentoMes  []FaturamentoPorMes     `json:"faturamentoPorMes"`
	Profissionais   []ProfissionalResumo    `json:"profissionais"`
}

type AgendamentosResumo struct {
	Hoje   int64 `json:"hoje"`
	Semana int64 `json:"semana"`
	Mes    int64 `json:"mes"`
}

type FaturamentoResumo struct {
	Hoje   float64 `json:"hoje"`
	Semana float64 `json:"semana"`
	Mes    float64 `json:"mes"`
}

type FaturamentoPorMes struct {
	Ano          int64   `json:"ano"`
	Mes          int64   `json:"mes"`
	Periodo      string  `json:"periodo"`
	Faturamento  float64 `json:"faturamento"`
}

type ProfissionalResumo struct {
	ProfissionalID  int64  `json:"profissional_id"`
	Nome            string `json:"nome"`
	Especialidade   string `json:"especialidade"`
	FotoURL         string `json:"foto_url"`
	TotalHoje       int64  `json:"totalHoje"`
	TotalSemana     int64  `json:"totalSemana"`
}