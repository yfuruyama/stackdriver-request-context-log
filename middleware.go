package stackdriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
)

// RequestLogging creates a middleware which logs a request log
func RequestLogging(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			before := time.Now()

			traceId := getTraceId(r)
			trace := fmt.Sprintf("projects/%s/traces/%s", config.ProjectId, traceId)

			contextLogger := &ContextLogger{
				out:            config.ContextLogOut,
				Trace:          trace,
				Severity:       config.Severity,
				loggedSeverity: make([]Severity, 0, 10),
			}
			ctx := context.WithValue(r.Context(), contextLoggerKey, contextLogger)
			r = r.WithContext(ctx)

			wrw := &wrappedResponseWriter{ResponseWriter: w}
			defer func() {
				// logging
				elapsed := time.Since(before)
				maxSeverity := contextLogger.maxSeverity()
				err := writeRequestLog(r, config, wrw.status, wrw.responseSize, elapsed, trace, maxSeverity)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				}
			}()
			next.ServeHTTP(wrw, r)
		}
		return http.HandlerFunc(fn)
	}
}

func getTraceId(r *http.Request) string {
	span := trace.FromContext(r.Context())
	if span != nil {
		return span.SpanContext().TraceID.String()
	}

	// there is no span yet, so create one

	httpFormat := &propagation.HTTPFormat{}
	if sc, ok := httpFormat.SpanContextFromRequest(r); ok {
		return sc.TraceID.String()
	} else {
		// TODO
		uniqueBytes := sha256.Sum256([]byte(time.Now().Format(time.RFC3339Nano)))
		return hex.EncodeToString(uniqueBytes[:16])
	}
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	status       int
	responseSize int
}

func (w *wrappedResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.responseSize += n
	return n, err
}

func writeRequestLog(r *http.Request, config *Config, status int, responseSize int, elapsed time.Duration, trace string, severity Severity) error {
	requestLog := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339Nano),
		"logging.googleapis.com/trace": trace,
		"severity":                     severity.String(),
		"httpRequest": map[string]interface{}{
			"requestMethod":                  r.Method,
			"requestUrl":                     r.URL.RequestURI(),
			"requestSize":                    fmt.Sprintf("%d", r.ContentLength),
			"status":                         status,
			"responseSize":                   fmt.Sprintf("%d", responseSize),
			"userAgent":                      r.UserAgent(),
			"remoteIp":                       r.RemoteAddr,
			"serverIp":                       getServerIp(),
			"referer":                        r.Referer(),
			"latency":                        fmt.Sprintf("%fs", elapsed.Seconds()),
			"cacheLookUp":                    false,
			"cacheHit":                       false,
			"cacheValidatedWithOriginServer": false,
			"protocol":                       r.Proto,
		},
		"logType": "request_log",
	}
	for k, v := range config.AdditionalFields {
		requestLog[k] = v
	}
	requestLogJson, err := json.Marshal(requestLog)
	if err != nil {
		return err
	}
	requestLogJson = append(requestLogJson, '\n')

	_, err = config.RequestLogOut.Write(requestLogJson)
	return err
}

func getServerIp() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return ""
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}
