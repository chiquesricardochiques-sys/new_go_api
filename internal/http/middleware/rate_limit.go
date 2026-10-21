package middleware

import "net/http"

// RateLimit é provisório.
// A limitação real de requisições será implementada depois.
func RateLimit(next http.Handler) http.Handler {
	return next
}
