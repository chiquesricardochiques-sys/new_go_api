
package middleware

import "net/http"

var origensPermitidas = map[string]bool{
	"http://localhost:8081":  true,
	"http://127.0.0.1:8081":  true,
	"http://localhost:8082":  true,
	"http://127.0.0.1:8082":  true,
	"http://localhost:19006": true,
	"http://127.0.0.1:19006": true,
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origensPermitidas[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Authorization, X-Request-ID, Accept",
			)
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			if w.Header().Get("Access-Control-Allow-Origin") == "" {
				http.Error(w, "CORS origin not allowed", http.StatusForbidden)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}