package salao

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
		"POST /saloes/get",
		handler.Get,
	)

	mux.HandleFunc(
		"POST /saloes/update",
		handler.Update,
	)
}