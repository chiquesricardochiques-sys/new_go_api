package salao

import "time"

type GetRequest struct {
	SalaoID int64 `json:"salao_id"`
}

type UpdateRequest struct {
	SalaoID    int64   `json:"salao_id"`
	Nome       *string `json:"nome"`
	Slug       *string `json:"slug"`
	Telefone   *string `json:"telefone"`
	Email      *string `json:"email"`
	Whatsapp   *string `json:"whatsapp"`
	Endereco   *string `json:"endereco"`
	Cidade     *string `json:"cidade"`
	Estado     *string `json:"estado"`
	CEP        *string `json:"cep"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	LogoURL    *string `json:"logo_url"`
	Status     *string `json:"status"`
}

type Salao struct {
	ID         int64      `json:"id"`
	Nome       string     `json:"nome"`
	Slug       string     `json:"slug"`
	Telefone   *string    `json:"telefone"`
	Email      *string    `json:"email"`
	Whatsapp   *string    `json:"whatsapp"`
	Endereco   *string    `json:"endereco"`
	Cidade     *string    `json:"cidade"`
	Estado     *string    `json:"estado"`
	CEP        *string    `json:"cep"`
	Latitude   *float64   `json:"latitude"`
	Longitude  *float64   `json:"longitude"`
	LogoURL    *string    `json:"logo_url"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}