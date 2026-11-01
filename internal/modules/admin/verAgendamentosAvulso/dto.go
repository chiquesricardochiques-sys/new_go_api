package verAgendamentosAvulso

type FiltroAgendamentosAvulsos struct {
    SalaoID int64  `json:"salao_id"`
    Data    string `json:"data"`
}

type CancelarAgendamentoRequest struct {
    SalaoID       int64 `json:"salao_id"`
    AgendamentoID int64 `json:"agendamento_id"`
}

type AgendamentoAvulso struct {
    ID              int64  `json:"id"`
    Data            string `json:"data"`
    Hora            string `json:"hora"`
    HoraFim         string `json:"hora_fim"`
    NomeCliente     string `json:"nome_cliente"`
    TelefoneCliente string `json:"telefone_cliente"`
    DuracaoMinutos  int    `json:"duracao_minutos"`
    Status          string `json:"status"`
    Observacao      string `json:"observacao"`
    Origem          string `json:"origem"`
}

type CancelarAgendamentoResponse struct {
    Sucesso bool   `json:"sucesso"`
    Mensagem string `json:"mensagem"`
}