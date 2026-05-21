package middleware

import "net/http"

func (m *Middleware) DeclareCorsMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", m.allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", m.allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", m.allowedHeaders)
			w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

			//"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}
}
