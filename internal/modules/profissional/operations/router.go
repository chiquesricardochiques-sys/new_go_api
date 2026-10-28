package profissional

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes registra as rotas sem lookup de clientes
// (a agenda devolve só cliente_id, sem nome/telefone).
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	RegisterRoutesComClientes(mux, db, nil)
}

// RegisterRoutesComClientes recebe uma implementação de ClienteLookup
// (ex.: chamando o Auth) para preencher nome/telefone/whatsapp na agenda.
func RegisterRoutesComClientes(mux *http.ServeMux, db *sql.DB, clientes ClienteLookup) {
	repository := NewRepository(db)
	service := NewService(repository, clientes)
	handler := NewHandler(service)

	// Substitui o antigo POST /login-profissional (login agora é no Auth).
	mux.HandleFunc("POST /api/profissional/me", comData(service.Me))

	mux.HandleFunc("POST /api/profissional/perfil", comData(service.Perfil))
	mux.HandleFunc("POST /api/profissional/agenda", comData(service.Agenda))
	mux.HandleFunc("POST /api/profissional/alterar-status", handler.AlterarStatus)
	mux.HandleFunc("POST /api/profissional/resumo", comData(service.Resumo))
	mux.HandleFunc("POST /api/profissional/historico", comData(service.Historico))
	mux.HandleFunc("POST /api/profissional/horarios-dia", comData(service.HorariosDia))
	mux.HandleFunc("POST /api/profissional/faturamento", comData(service.Faturamento))
}