
package professionals

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

	mux.HandleFunc("POST /professionals/list", handler.List)
	mux.HandleFunc("POST /professionals/update", handler.Update)
}
