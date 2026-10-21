package auth

type LoginRequest struct {
	IDSalao int64  `json:"id_salao"`
	Email   string `json:"email"`
	Senha   string `json:"senha"`
}

type AdminResponse struct {
	ID       int64  `json:"id"`
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	CriadoEm string `json:"criado_em"`
}

type ProfissionalResponse struct {
	ID            int64  `json:"id"`
	Nome          string `json:"nome"`
	Email         string `json:"email"`
	Especialidade string `json:"especialidade"`
	Img           string `json:"img"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}
