package stackdriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"go.opencensus.io/exporter/stackdriver/propagation"
)

func Handler(logger *Logger, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		before := time.Now()
		wrw := &WrappedResponseWriter{ResponseWriter: w}

		traceId := getTraceId(r)
		ctx := context.WithValue(r.Context(), "traceId", traceId) // TODO

		trace := fmt.Sprintf("projects/%s/traces/%s", logger.projectId, traceId)
		contextLogger := logger.appLogger.WithField("trace", trace)
		ctx = context.WithValue(ctx, "contextLogger", contextLogger) // TODO

		r = r.WithContext(ctx)

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

func getTraceId(r *http.Request) string {
	httpFormat := &propagation.HTTPFormat{}
	if sc, ok := httpFormat.SpanContextFromRequest(r); ok {
		return sc.TraceID.String()
	} else {
		// TODO
		uniqueBytes := sha256.Sum256([]byte(time.Now().Format(time.RFC3339Nano)))
		return hex.EncodeToString(uniqueBytes[:16])
	}
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