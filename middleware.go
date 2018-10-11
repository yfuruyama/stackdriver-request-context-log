package stackdriver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
)

// RequestLogging creates the middleware which logs a request log and creates a request-context logger
func RequestLogging(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			before := time.Now()

			traceId := getTraceId(r)
			if traceId == "" {
				// there is no span yet, so create one
				var ctx context.Context
				traceId, ctx = generateTraceId(r)
				r = r.WithContext(ctx)
			}

			trace := fmt.Sprintf("projects/%s/traces/%s", config.ProjectId, traceId)

			contextLogger := &ContextLogger{
				out:            config.ContextLogOut,
				Trace:          trace,
				Severity:       config.Severity,
				AdditionalData: config.AdditionalData,
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

	httpFormat := &propagation.HTTPFormat{}
	if sc, ok := httpFormat.SpanContextFromRequest(r); ok {
		return sc.TraceID.String()
	}

	return ""
}

func generateTraceId(r *http.Request) (string, context.Context) {
	ctx, span := trace.StartSpan(r.Context(), "")
	sc := span.SpanContext()
	return sc.TraceID.String(), ctx
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

type HttpRequest struct {
	RequestMethod                  string `json:"requestMethod"`
	RequestUrl                     string `json:"requestUrl"`
	RequestSize                    string `json:"requestSize"`
	Status                         int    `json:"status"`
	ResponseSize                   string `json:"responseSize"`
	UserAgent                      string `json:"userAgent"`
	RemoteIp                       string `json:"remoteIp"`
	ServerIp                       string `json:"serverIp"`
	Referer                        string `json:"referer"`
	Latency                        string `json:"latency"`
	CacheLookup                    bool   `json:"cacheLookup"`
	CacheHit                       bool   `json:"cacheHit"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer"`
	Protocol                       string `json:"protocol"`
}

type HttpRequestLog struct {
	Time           string         `json:"time"`
	Trace          string         `json:"logging.googleapis.com/trace"`
	Severity       string         `json:"severity"`
	HttpRequest    HttpRequest    `json:"httpRequest"`
	AdditionalData AdditionalData `json:"data,omitempty"`
}

func writeRequestLog(r *http.Request, config *Config, status int, responseSize int, elapsed time.Duration, trace string, severity Severity) error {
	requestLog := &HttpRequestLog{
		Time:     time.Now().Format(time.RFC3339Nano),
		Trace:    trace,
		Severity: severity.String(),
		HttpRequest: HttpRequest{
			RequestMethod:                  r.Method,
			RequestUrl:                     r.URL.RequestURI(),
			RequestSize:                    fmt.Sprintf("%d", r.ContentLength),
			Status:                         status,
			ResponseSize:                   fmt.Sprintf("%d", responseSize),
			UserAgent:                      r.UserAgent(),
			RemoteIp:                       getRemoteIp(r),
			ServerIp:                       getServerIp(),
			Referer:                        r.Referer(),
			Latency:                        fmt.Sprintf("%fs", elapsed.Seconds()),
			CacheLookup:                    false,
			CacheHit:                       false,
			CacheValidatedWithOriginServer: false,
			Protocol:                       r.Proto,
		},
		AdditionalData: config.AdditionalData,
	}
	requestLogJson, err := json.Marshal(requestLog)
	if err != nil {
		return err
	}
	requestLogJson = append(requestLogJson, '\n')

	_, err = config.RequestLogOut.Write(requestLogJson)
	return err
}

func getRemoteIp(r *http.Request) string {
	parts := strings.Split(r.RemoteAddr, ":")
	return parts[0]
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
