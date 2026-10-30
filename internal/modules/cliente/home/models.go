package home

import "time"

type HomeRequest struct {
	SalaoID int64 `json:"salao_id"`
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

type HomeData struct {
	Textos       map[string]string `json:"textos"`
	Salao        *Salao            `json:"salao"`
	Profissionais []Profissional   `json:"profissionais"`
}

type HomeResponse struct {
	Success bool      `json:"success"`
	Data    *HomeData `json:"data,omitempty"`
}

type salaonDB struct {
	ID         int64
	Nome       string
	Telefone   *string
	Email      *string
	Whatsapp   *string
	Endereco   *string
	Cidade     *string
	Estado     *string
	CEP        *string
	Latitude   *float64
	Longitude  *float64
	LogoURL    *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}