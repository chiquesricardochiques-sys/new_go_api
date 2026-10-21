package client_home

import "time"

type ClientHomeRequest struct {
	SalaoID  int64 `json:"salao_id"`
	IDCliente int64 `json:"id_cliente"`
}

type Texto struct {
	Chave    string `json:"chave"`
	Conteudo string `json:"conteudo"`
}

type Salao struct {
	ID        int64    `json:"id"`
	Nome      string   `json:"nome"`
	Telefone  *string  `json:"telefone"`
	Email     *string  `json:"email"`
	Whatsapp  *string  `json:"whatsapp"`
	Endereco  *string  `json:"endereco"`
	Cidade    *string  `json:"cidade"`
	Estado    *string  `json:"estado"`
	CEP       *string  `json:"cep"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	LogoURL   *string  `json:"logo_url"`
}

type Profissional struct {
	ID            int64   `json:"id"`
	Nome          string  `json:"nome"`
	Especialidade *string `json:"especialidade"`
	FotoURL       *string `json:"foto_url"`
}

type ProximoAgendamento struct {
	ID           int64     `json:"id"`
	Inicio       time.Time `json:"inicio"`
	Fim          time.Time `json:"fim"`
	Status       string    `json:"status"`
	Profissional string    `json:"profissional"`
}

type ClientHomeData struct {
	Textos             map[string]string  `json:"textos"`
	Salao              *Salao             `json:"salao"`
	Profissionais      []Profissional     `json:"profissionais"`
	ProximoAgendamento *ProximoAgendamento `json:"proximo_agendamento"`
}

type ClientHomeResponse struct {
	Success bool            `json:"success"`
	Data    *ClientHomeData `json:"data,omitempty"`
}