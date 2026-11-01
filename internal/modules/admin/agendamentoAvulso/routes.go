
package agendamentoavulso

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	const prefix = "/admin/agendamento-avulso"

	// Listar profissionais do salão
	mux.HandleFunc(
		"POST "+prefix+"/profissionais",
		handler.Profissionais,
	)

	// Buscar serviços disponíveis para o profissional
	mux.HandleFunc(
		"POST "+prefix+"/servicos-profissional",
		handler.ServicosProfissional,
	)

	// Buscar horários disponíveis
	mux.HandleFunc(
		"POST "+prefix+"/horarios-disponiveis",
		handler.HorariosDisponiveis,
	)

	// Carregar dados completos da tela
	mux.HandleFunc(
		"POST "+prefix+"/schedule-page",
		handler.SchedulePage,
	)

	// Criar agendamento avulso
	mux.HandleFunc(
		"POST "+prefix+"/agendar",
		handler.Agendar,
	)
}
