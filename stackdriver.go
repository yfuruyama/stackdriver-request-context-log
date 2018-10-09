package stackdriver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type AdditionalFields map[string]interface{}

type Logger struct {
	requestLogger *log.Logger
	appLogger     *log.Logger
	projectId     string
	additional    AdditionalFields
	// Level: int
}

func NewLogger(outRequestLog io.Writer, outAppLog io.Writer, projectId string, additional AdditionalFields) *Logger {
	requestLogger := log.New(outRequestLog, "", 0)
	appLogger := log.New(outAppLog, "", 0)

	return &Logger{
		requestLogger: requestLogger,
		appLogger:     appLogger,
		projectId:     projectId,
		additional:    additional,
	}
}

func (l *Logger) WriteRequestLog(r *http.Request, status int, responseSize int, elapsed time.Duration, severity Severity) error {
	traceId := r.Context().Value("traceId").(string)
	trace := fmt.Sprintf("projects/%s/traces/%s", l.projectId, traceId)

	latency := fmt.Sprintf("%fs", elapsed.Seconds())

	requestLog := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339Nano),
		"logging.googleapis.com/trace": trace,
		"severity":                     severity.String(),
		"httpRequest": map[string]interface{}{
			"requestMethod":                  r.Method,
			"requestUrl":                     r.URL.Path,
			"requestSize":                    fmt.Sprintf("%d", r.ContentLength),
			"status":                         status,
			"responseSize":                   fmt.Sprintf("%d", responseSize),
			"userAgent":                      r.UserAgent(),
			"remoteIp":                       r.RemoteAddr,
			"serverIp":                       "localhost",
			"referer":                        r.Referer(),
			"latency":                        latency,
			"cacheLookUp":                    false, // TODO
			"cacheHit":                       false, // TODO
			"cacheValidatedWithOriginServer": false, // TODO
			"protocol":                       r.Proto,
		},
		// "logType": "request_log",
	}
	for k, v := range l.additional {
		requestLog[k] = v
	}
	requestLogJson, err := json.Marshal(requestLog)
	if err != nil {
		return err
	}

	l.requestLogger.Println(string(requestLogJson))

	return nil
}

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
type Severity int

const (
	SeverityDefault   Severity = 0
	SeverityDebug     Severity = 100
	SeverityInfo      Severity = 200
	SeverityNotice    Severity = 300
	SeverityWarning   Severity = 400
	SeverityError     Severity = 500
	SeverityCritical  Severity = 600
	SeverityAlert     Severity = 700
	SeverityEmergency Severity = 800
)

func (s Severity) String() string {
	switch s {
	case SeverityDefault:
		return "DEFAULT"
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityNotice:
		return "NOTICE"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	case SeverityAlert:
		return "ALERT"
	case SeverityEmergency:
		return "EMERGENCY"
	}
	return "UNKNOWN"
}

type AppLogger struct {
	logger         *log.Logger
	Trace          string
	Severity       Severity
	loggedSeverity []Severity
}

func LoggerFromRequest(r *http.Request) *AppLogger {
	return r.Context().Value("appLogger").(*AppLogger)
}

func (a *AppLogger) Infof(format string, args ...interface{}) {
	a.log(SeverityInfo, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Warnf(format string, args ...interface{}) {
	a.log(SeverityWarning, fmt.Sprintf(format, args...))
}

func (a *AppLogger) log(severity Severity, msg string) {
	if severity < a.Severity {
		return
	}

	a.loggedSeverity = append(a.loggedSeverity, severity)

	appLog := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339Nano),
		"logging.googleapis.com/trace": a.Trace,
		"severity":                     severity.String(),
		"message":                      msg,
		// "logType":                      "app_log",
	}
	b, err := json.Marshal(appLog)
	if err != nil {
		panic(err) // TODO
	}

	a.logger.Println(string(b))
}

func (a *AppLogger) maxSeverity() Severity {
	max := SeverityDefault
	if len(a.loggedSeverity) == 0 {
		return max
	}

	for _, s := range a.loggedSeverity {
		if s > max {
			max = s
		}
	}

	return max
}
