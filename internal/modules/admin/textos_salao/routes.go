package textos_salao

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
		"POST /saloes/textos/list",
		handler.List,
	)

	mux.HandleFunc(
		"POST /saloes/textos/get",
		handler.Get,
	)

	mux.HandleFunc(
		"POST /saloes/textos/add",
		handler.Add,
	)

	mux.HandleFunc(
		"POST /saloes/textos/update",
		handler.Update,
	)

	mux.HandleFunc(
		"POST /saloes/textos/delete",
		handler.Delete,
	)
}