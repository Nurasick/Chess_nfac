package middleware

import (
	"net/http"
	"strings"

	"github.com/chess-nfac/backend/config"
)

func isLocalhostOrigin(origin string) bool {
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:")
}

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	allowedOrigins := make(map[string]bool, len(cfg.CORSOrigins))
	for _, o := range cfg.CORSOrigins {
		allowedOrigins[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && (len(allowedOrigins) == 0 || allowedOrigins[origin] || isLocalhostOrigin(origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")

			if strings.ToUpper(r.Method) == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
