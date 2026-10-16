package auth

import "net/http"

type Middleware struct {
	required         bool
	allowedPrincipal string
}

func New(required bool, allowedPrincipal string) Middleware {
	return Middleware{required: required, allowedPrincipal: allowedPrincipal}
}

func (m Middleware) Wrap(next http.Handler) http.Handler {
	if !m.required {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		principal := r.Header.Get("X-MS-CLIENT-PRINCIPAL-ID")
		if principal == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if m.allowedPrincipal != "" && principal != m.allowedPrincipal {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
