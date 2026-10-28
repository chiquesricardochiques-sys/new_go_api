package statusAG

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository, db)
	handler := NewHandler(service)

	mux.HandleFunc("POST /api/profissional/iniciar-corte", handler.IniciarCorte)
	mux.HandleFunc("POST /api/profissional/finalizar-corte", handler.FinalizarCorte)
	mux.HandleFunc("POST /api/profissional/cancelar-agendamento", handler.CancelarAgendamento)
	mux.HandleFunc("POST /api/profissional/nao-compareceu", handler.NaoCompareceu)
}