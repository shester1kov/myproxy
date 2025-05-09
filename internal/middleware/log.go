package middleware

import (
	"log"
	"net/http"
	"time"
)

// middleware для логирования поступающих запросов
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("Начало обработки %s %s\n", r.Method, r.URL.Path)

			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			log.Printf(
				"Завершение обработки %s %s | Статус: %d | Длитеьльность: %s",
				r.Method,
				r.URL.Path,
				rw.status,
				duration,
			)
		},
	)
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
