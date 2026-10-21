package salary

type Request struct {
	SalaoID        int64  `json:"salao_id"`
	ProfissionalID int64  `json:"profissional_id"`
	MesAno         string `json:"mes_ano"`
}

type Periodo struct {
	Qtd         int     `json:"qtd"`
	TotalGerado float64 `json:"totalGerado"`
	ValorGanho  float64 `json:"valorGanho"`
}

type ResponseData struct {
	Nome               string  `json:"nome"`
	FotoURL            *string `json:"foto_url"`
	TipoRemuneracao    string  `json:"tipo_remuneracao"`
	PercentualComissao float64 `json:"percentual_comissao"`
	SalarioFixo        float64 `json:"salario_fixo"`
	Hoje               Periodo `json:"hoje"`
	Semana             Periodo `json:"semana"`
	MesAtual           Periodo `json:"mesAtual"`
	MesFiltro          Periodo `json:"mesFiltro"`
}

type AppointmentEarning struct {
	ID    int64
	Data  string
	Total float64
}
