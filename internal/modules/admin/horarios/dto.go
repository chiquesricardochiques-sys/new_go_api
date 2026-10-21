package horarios

type ListRequest struct {
	IDSalao        int64  `json:"id_salao"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
}

type CreateRequest struct {
	IDSalao        int64  `json:"id_salao"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
	Hora           string `json:"hora"`
}

type DeleteRequest struct {
	IDSalao int64 `json:"id_salao"`
	ID      int64 `json:"id"`
}

type Horario struct {
	ID             int64  `json:"id"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
	Hora           string `json:"hora"`
	Status         string `json:"status"`
}
