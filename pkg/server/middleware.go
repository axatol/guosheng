package server

import (
	"net/http"
)

func middlewareContentType(defaultContentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			if existing := r.Header.Get("Content-Type"); existing != "" {
				return
			}

			r.Header.Set("Content-Type", defaultContentType)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
