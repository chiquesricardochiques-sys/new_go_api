package avisos

import "time"

type ListarAvisosRequest struct {
	SalaoID  int64 `json:"salao_id"`
	Historico bool `json:"historico"`
}

type Aviso struct {
	ID        int64     `json:"id"`
	SalaoID   int64     `json:"salao_id"`
	Titulo    string    `json:"titulo"`
	Mensagem  string    `json:"mensagem"`
	Tipo      string    `json:"tipo"`
	Lido      bool      `json:"lido"`
	CreatedAt time.Time `json:"created_at"`
}

type ListarAvisosResponse struct {
	Avisos []Aviso `json:"avisos"`
	Total  int     `json:"total"`
}

type ApiResponse struct {
	Success bool                  `json:"success"`
	Data    *ListarAvisosResponse `json:"data,omitempty"`
	Message string                `json:"message,omitempty"`
}