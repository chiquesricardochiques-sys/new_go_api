package client_home

type ClientHomeRequest struct {
	SalaoID    int64 `json:"salao_id"`
	AuthUserID int64 `json:"auth_user_id"`
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

// ProximoAgendamento: data/hora já separadas (horário do salão, sem fuso)
// e com os serviços, que é o que o cartão de status mostra.
type ProximoAgendamento struct {
	ID           int64    `json:"id"`
	Data         string   `json:"data"`     // YYYY-MM-DD
	Hora         string   `json:"hora"`     // HH:MM
	HoraFim      string   `json:"hora_fim"` // HH:MM
	Status       string   `json:"status"`
	Profissional string   `json:"profissional"`
	Servicos     []string `json:"servicos"`
	DuracaoTotal int      `json:"duracao_total"`
	PrecoTotal   float64  `json:"preco_total"`
}

type ClientHomeData struct {
	Textos             map[string]string   `json:"textos"`
	Salao              *Salao              `json:"salao"`
	Profissionais      []Profissional      `json:"profissionais"`
	ProximoAgendamento *ProximoAgendamento `json:"proximo_agendamento"`
}

type ClientHomeResponse struct {
	Success bool            `json:"success"`
	Data    *ClientHomeData `json:"data,omitempty"`
}