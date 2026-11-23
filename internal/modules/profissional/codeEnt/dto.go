package codeEnt

type BuscarSalaoRequest struct {
	CodigoPainel string `json:"codigo_painel"`
}

type SalaoResponse struct {
	ID     int64  `json:"id"`
	Nome   string `json:"nome"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

type BuscarSalaoResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Salao   *SalaoResponse `json:"salao,omitempty"`
}