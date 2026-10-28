package gradehorarios

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)

	// Grade do dia já com livre / agendado / bloqueado em cada horário.
	mux.HandleFunc("POST /api/profissional/grade/horarios-dia",
		comData("", service.HorariosDia))

	// Bloqueia um horário LIVRE (se tiver agendamento, recusa com 409).
	mux.HandleFunc("POST /api/profissional/grade/bloquear",
		comData("Horário bloqueado com sucesso!", service.Bloquear))

	// Remove um bloqueio.
	mux.HandleFunc("POST /api/profissional/grade/desbloquear",
		comData("Horário desbloqueado com sucesso!", service.Desbloquear))
}