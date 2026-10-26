package schedule

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(
	mux *http.ServeMux,
	db *sql.DB,
) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc(
		"POST /horarios-disponiveis",
		handler.HorariosDisponiveis,
	)

	mux.HandleFunc(
		"POST /schedule-page",
		handler.SchedulePage,
	)

	mux.HandleFunc(
		"POST /servicos-profissional",
		handler.ServicosProfissional,
	)

	mux.HandleFunc(
		"POST /agendar",
		handler.Agendar,
	)
}