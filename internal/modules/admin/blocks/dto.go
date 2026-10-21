package blocks

type ListRequest struct {
	IDSalao        int64  `json:"id_salao"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
}

type AddDayRequest struct {
	IDSalao        int64  `json:"id_salao"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
	Motivo         string `json:"motivo"`
}

type AddHourRequest struct {
	IDSalao        int64  `json:"id_salao"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
	Hora           string `json:"hora"`
	Motivo         string `json:"motivo"`
}

type RemoveRequest struct {
	IDSalao int64 `json:"id_salao"`
	ID      int64 `json:"id"`
}

type DiaBloqueado struct {
	ID             int64  `json:"id"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
	Motivo         string `json:"motivo"`
}

type HoraBloqueada struct {
	ID             int64  `json:"id"`
	ProfissionalID int64  `json:"profissional_id"`
	Data           string `json:"data"`
	Hora           string `json:"hora"`
	Motivo         string `json:"motivo"`
}

type ListResponse struct {
	DiasBloqueados  []DiaBloqueado  `json:"diasBloqueados"`
	HorasBloqueadas []HoraBloqueada `json:"horasBloqueadas"`
	Data            string          `json:"data"`
}
