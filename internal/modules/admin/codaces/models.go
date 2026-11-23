package codaces

type CodigoRequest struct {
	Codigo string `json:"codigo"`
}

type CodigoResponse struct {
	SalaoID uint64 `json:"salao_id"`
	Codigo   string `json:"codigo"`
}

type MensagemResponse struct {
	Mensagem string `json:"mensagem"`
}

type ErrorResponse struct {
	Erro string `json:"erro"`
}