package statusAG

type BaseRequest struct {
	IDSalao        int64 `json:"id_salao"`
	ProfissionalID int64 `json:"profissional_id"`
}

type IniciarCorteRequest struct {
	BaseRequest
	AgendamentoID int64 `json:"agendamento_id"`
}

type FinalizarCorteRequest struct {
	BaseRequest
	AgendamentoID int64 `json:"agendamento_id"`
}

type CancelarAgendamentoRequest struct {
	BaseRequest
	AgendamentoID int64  `json:"agendamento_id"`
	Motivo        string `json:"motivo"`
}

type NaoCompareceuRequest struct {
	BaseRequest
	AgendamentoID int64 `json:"agendamento_id"`
}

type StatusResponse struct {
	NovoStatus string `json:"novo_status"`
}