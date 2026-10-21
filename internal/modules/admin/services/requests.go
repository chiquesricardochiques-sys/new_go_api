package services

// ============================================================
// SERVIÇOS GLOBAIS
// ============================================================

type ListGlobalServicesRequest struct{}

// ============================================================
// SERVIÇOS DO SALÃO
// ============================================================

type ListSalonServicesRequest struct {
	SalaoID uint64 `json:"salao_id"`
}

type AddSalonServiceRequest struct {
	SalaoID        uint64  `json:"salao_id"`
	ServicoID      uint64  `json:"servico_id"`
	Preco          float64 `json:"preco"`
	DuracaoMinutos uint     `json:"duracao_minutos"`
}

type DeactivateSalonServiceRequest struct {
	SalaoID   uint64 `json:"salao_id"`
	ServicoID uint64 `json:"servico_id"`
}

type UnlinkSalonServiceRequest struct {
	SalaoID   uint64 `json:"salao_id"`
	ServicoID uint64 `json:"servico_id"`
}

// Mantido para busca específica de um serviço dentro do salão.
type GetServiceRequest struct {
	SalaoID   uint64 `json:"salao_id"`
	ServicoID uint64 `json:"servico_id"`
}

// ============================================================
// PROFISSIONAIS
// ============================================================

type ListProfessionalsRequest struct {
	SalaoID uint64 `json:"salao_id"`
}

type ListProfessionalsByServiceRequest struct {
	SalaoID   uint64 `json:"salao_id"`
	ServicoID uint64 `json:"servico_id"`
}

// ============================================================
// SERVIÇOS DO PROFISSIONAL
// ============================================================

// Lista TODOS os serviços ativos do salão,
// indicando quais estão selecionados para o profissional.
type ListProfessionalServicesRequest struct {
	SalaoID        uint64 `json:"salao_id"`
	ProfissionalID uint64 `json:"profissional_id"`
}

// Lista somente os serviços efetivamente vinculados
// ao profissional.
type ListLinkedProfessionalServicesRequest struct {
	SalaoID        uint64 `json:"salao_id"`
	ProfissionalID uint64 `json:"profissional_id"`
}

type AddProfessionalServiceRequest struct {
	SalaoID        uint64 `json:"salao_id"`
	ProfissionalID uint64 `json:"profissional_id"`
	ServicoID      uint64 `json:"servico_id"`
}

type RemoveProfessionalServiceRequest struct {
	SalaoID        uint64 `json:"salao_id"`
	ProfissionalID uint64 `json:"profissional_id"`
	ServicoID      uint64 `json:"servico_id"`
}