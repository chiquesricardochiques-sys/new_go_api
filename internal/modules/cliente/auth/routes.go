package cliente

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
		"POST /auth/cliente/verificar",
		handler.Verificar,
	)
	mux.HandleFunc(
		"POST /auth/cliente/adicionar",
		handler.Adicionar,
	)
}