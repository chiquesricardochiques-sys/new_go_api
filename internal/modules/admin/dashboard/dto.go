package dashboard

type ResumoResponse struct {
	TotalHoje        int64                `json:"totalHoje"`
	TotalSemana      int64                `json:"totalSemana"`
	TotalMes         int64                `json:"totalMes"`
	CanceladosHoje   int64                `json:"canceladosHoje"`
	CanceladosSemana int64                `json:"canceladosSemana"`
	PorProfissional  []ProfissionalResumo `json:"porProfissional"`
}

type ProfissionalResumo struct {
	ProfissionalID int64  `json:"profissional_id"`
	Nome           string `json:"nome"`
	TotalHoje      int64  `json:"totalHoje"`
	TotalSemana    int64  `json:"totalSemana"`
}