package cliente

type RelacaoRequest struct {
	AuthUserID int64 `json:"auth_user_id"`
	SalaoID    int64 `json:"salao_id"`
}

type RelacaoResponse struct {
	IDCliente int64 `json:"id_cliente"`
}