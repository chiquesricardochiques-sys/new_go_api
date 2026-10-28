
package profissional

import (
	"database/sql"
	"net/http"
)

// RegisterRoutes registra as rotas do módulo profissional.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc(
		"POST /api/profissional/profissional/addRelations",
		handler.AddRelations,
	)

	mux.HandleFunc(
		"POST /api/profissional/profissional/verify",
		handler.Verify,
	)
}
