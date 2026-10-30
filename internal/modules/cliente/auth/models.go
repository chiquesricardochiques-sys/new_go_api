package cliente

// RelacaoRequest: o telefone é opcional e pode vir com máscara,
// por exemplo "(19) 99999-8888". O service normaliza antes de gravar.
type RelacaoRequest struct {
	AuthUserID int64  `json:"auth_user_id"`
	SalaoID    int64  `json:"salao_id"`
	Telefone   string `json:"telefone"`
}

type RelacaoResponse struct {
	IDCliente int64  `json:"id_cliente"`
	Telefone  string `json:"telefone,omitempty"`
}