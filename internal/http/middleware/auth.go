package middleware

import "net/http"

// Auth ainda não valida tokens.
// A validação será implementada após confirmarmos
// o contrato de autenticação da API AUTH.
func Auth(next http.Handler) http.Handler {
	return next
}
