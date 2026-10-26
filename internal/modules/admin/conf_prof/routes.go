package conf_prof

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(
	mux *http.ServeMux,
	db *sql.DB,
) {
	repository := NewRepository(db)

	service := NewService(
		repository,
	)

	handler := NewHandler(
		service,
	)

	// =====================================================
	// ADMIN
	// =====================================================

	mux.HandleFunc(
		"GET /admin/configuracao-profissional/rendimento/{profissionalId}",
		handler.BuscarPodeVerRendimento,
	)

	mux.HandleFunc(
		"PUT /admin/configuracao-profissional/rendimento/{profissionalId}",
		handler.DefinirPodeVerRendimento,
	)

	// =====================================================
	// PROFISSIONAL
	// =====================================================

	mux.HandleFunc(
		"GET /profissional/configuracoes/rendimento",
		handler.BuscarPodeVerRendimentoProfissional,
	)
}