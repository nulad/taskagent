package middleware

import (
	"net/http"
)

// CORSMiddleware returns middleware that enforces CORS for configured origins.
// Requests without an Origin header are passed through unchanged.
// Allowed origins receive proper CORS headers; disallowed origins receive 403.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !isOriginAllowed(origin, allowedOrigins) {
				writeJSONError(w, http.StatusForbidden, "origin not allowed")
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")

			// Handle preflight OPTIONS requests
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethodsHeader())
				w.Header().Set("Access-Control-Allow-Headers", allowedHeadersHeader())
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// For actual (non-OPTIONS) requests, set CORS headers and continue
			w.Header().Set("Access-Control-Allow-Methods", allowedMethodsHeader())
			w.Header().Set("Access-Control-Allow-Headers", allowedHeadersHeader())
			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func allowedMethodsHeader() string {
	return "GET, POST, PUT, PATCH, DELETE"
}

func allowedHeadersHeader() string {
	return "X-API-Key, Content-Type"
}
