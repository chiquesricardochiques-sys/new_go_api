package conf_prof

type Configuracao struct {
	ProfissionalID int64  `json:"profissional_id"`
	Chave          string `json:"chave"`
	Valor          any    `json:"valor"`
}

type PodeVerRendimento struct {
	ProfissionalID int64 `json:"profissional_id"`
	Permitido      bool  `json:"permitido"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}