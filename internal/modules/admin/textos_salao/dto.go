package textos_salao

import "time"

type TextoSalao struct {
	ID        int64     `json:"id"`
	SalaoID   int64     `json:"salao_id"`
	Chave     string    `json:"chave"`
	Conteudo  string    `json:"conteudo"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListRequest struct {
	SalaoID int64 `json:"salao_id"`
}

type GetRequest struct {
	SalaoID int64  `json:"salao_id"`
	Chave   string `json:"chave"`
}

type AddRequest struct {
	SalaoID  int64  `json:"salao_id"`
	Chave    string `json:"chave"`
	Conteudo string `json:"conteudo"`
}

type UpdateRequest struct {
	SalaoID  int64  `json:"salao_id"`
	Chave    string `json:"chave"`
	Conteudo string `json:"conteudo"`
}

type DeleteRequest struct {
	SalaoID int64  `json:"salao_id"`
	Chave   string `json:"chave"`
}