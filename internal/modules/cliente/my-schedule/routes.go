package myschedule

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc("POST /my-schedules", handler.MeusAgendamentos)
	mux.HandleFunc("POST /historico-agendamentos", handler.Historico)
	mux.HandleFunc("POST /deletar-agendamento", handler.Cancelar)
}