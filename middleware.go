package stackdriver

import (
	"net/http"
	"time"
)

func Handler(logger *Logger, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		before := time.Now()
		wrw := &WrappedResponseWriter{ResponseWriter: w}
		defer func() {
			// logging
			after := time.Since(before)
			err := logger.WriteRequestLog(r, wrw.status, wrw.responseSize, after)
			if err != nil {
				panic(err)
			}
		}()
		next.ServeHTTP(wrw, r)
	}
	return http.HandlerFunc(fn)
}

type WrappedResponseWriter struct {
	http.ResponseWriter
	status       int
	responseSize int
}

func (w *WrappedResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *WrappedResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.responseSize += n
	return n, err
}
