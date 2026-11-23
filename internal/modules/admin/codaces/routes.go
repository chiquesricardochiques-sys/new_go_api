package codaces

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
	controller := NewController(service)
	handler := NewHandler(controller)

	// --------------------------------------------------------
	// BUSCAR CÓDIGO DO SALÃO
	// --------------------------------------------------------

	mux.HandleFunc(
		"GET /admin/saloes/{salao_id}/codigo",
		handler.BuscarCodigo,
	)

	// --------------------------------------------------------
	// INSERIR CÓDIGO
	// --------------------------------------------------------

	mux.HandleFunc(
		"POST /admin/saloes/{salao_id}/codigo",
		handler.InserirCodigo,
	)

	// --------------------------------------------------------
	// ALTERAR CÓDIGO
	// --------------------------------------------------------

	mux.HandleFunc(
		"PUT /admin/saloes/{salao_id}/codigo",
		handler.AlterarCodigo,
	)
}