package professional_schedule

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
		db,
	)

	handler := NewHandler(service)

	mux.HandleFunc(
		"POST /professionals/schedule/get",
		handler.Get,
	)

	mux.HandleFunc(
		"POST /professionals/schedule/update",
		handler.Update,
	)

	mux.HandleFunc(
		"POST /professionals/schedule/reset",
		handler.Reset,
	)
}