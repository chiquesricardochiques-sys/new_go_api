
package admins

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)

	mux.HandleFunc(
		"POST /api/admin/admins/addRelations",
		handler.AddRelations,
	)

	mux.HandleFunc(
		"POST /api/admin/admins/verify",
		handler.Verify,
	)
}