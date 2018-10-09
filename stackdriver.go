package stackdriver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type AdditionalFields map[string]interface{}

type Logger struct {
	requestLogger *log.Logger
	appLogger     *log.Logger
	projectId     string
	additional    AdditionalFields
	severity      Severity
}

type Option func(*Logger)

func WithSeverity(severity Severity) Option {
	return func(l *Logger) {
		l.severity = severity
	}
}

func WithAdditionalFields(fields AdditionalFields) Option {
	return func(l *Logger) {
		l.additional = fields
	}
}

func WithOut(outRequestLog io.Writer, outAppLog io.Writer) Option {
	return func(l *Logger) {
		l.requestLogger = log.New(outRequestLog, "", 0)
		l.appLogger = log.New(outAppLog, "", 0)
	}
}

func NewLogger(projectId string, options ...Option) *Logger {
	logger := &Logger{
		projectId:     projectId,
		severity:      SeverityInfo,
		requestLogger: log.New(os.Stderr, "", 0),
		appLogger:     log.New(os.Stdout, "", 0),
	}
	for _, option := range options {
		option(logger)
	}
	return logger
}

func (l *Logger) WriteRequestLog(r *http.Request, status int, responseSize int, elapsed time.Duration, trace string, severity Severity) error {
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
		"logType": "request_log",
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
	return r.Context().Value(appLoggerKey).(*AppLogger)
}

func (a *AppLogger) Default(args ...interface{}) {
	a.log(SeverityDefault, fmt.Sprint(args...))
}

func (a *AppLogger) Defaultf(format string, args ...interface{}) {
	a.log(SeverityDefault, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Defaultln(args ...interface{}) {
	a.log(SeverityDefault, fmt.Sprintln(args...))
}

func (a *AppLogger) Debug(args ...interface{}) {
	a.log(SeverityDebug, fmt.Sprint(args...))
}

func (a *AppLogger) Debugf(format string, args ...interface{}) {
	a.log(SeverityDebug, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Debugln(args ...interface{}) {
	a.log(SeverityDebug, fmt.Sprintln(args...))
}

func (a *AppLogger) Info(args ...interface{}) {
	a.log(SeverityInfo, fmt.Sprint(args...))
}

func (a *AppLogger) Infof(format string, args ...interface{}) {
	a.log(SeverityInfo, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Infoln(args ...interface{}) {
	a.log(SeverityInfo, fmt.Sprintln(args...))
}

func (a *AppLogger) Notice(args ...interface{}) {
	a.log(SeverityNotice, fmt.Sprint(args...))
}

func (a *AppLogger) Noticef(format string, args ...interface{}) {
	a.log(SeverityNotice, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Noticeln(args ...interface{}) {
	a.log(SeverityNotice, fmt.Sprintln(args...))
}

func (a *AppLogger) Warning(args ...interface{}) {
	a.log(SeverityWarning, fmt.Sprint(args...))
}

func (a *AppLogger) Warningf(format string, args ...interface{}) {
	a.log(SeverityWarning, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Warningln(args ...interface{}) {
	a.log(SeverityWarning, fmt.Sprintln(args...))
}

func (a *AppLogger) Warn(args ...interface{}) {
	a.Warning(args...)
}

func (a *AppLogger) Warnf(format string, args ...interface{}) {
	a.Warningf(format, args...)
}

func (a *AppLogger) Warnln(args ...interface{}) {
	a.Warningln(args...)
}

func (a *AppLogger) Error(args ...interface{}) {
	a.log(SeverityError, fmt.Sprint(args...))
}

func (a *AppLogger) Errorf(format string, args ...interface{}) {
	a.log(SeverityError, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Errorln(args ...interface{}) {
	a.log(SeverityError, fmt.Sprintln(args...))
}

func (a *AppLogger) Critical(args ...interface{}) {
	a.log(SeverityCritical, fmt.Sprint(args...))
}

func (a *AppLogger) Criticalf(format string, args ...interface{}) {
	a.log(SeverityCritical, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Criticalln(args ...interface{}) {
	a.log(SeverityCritical, fmt.Sprintln(args...))
}

func (a *AppLogger) Alert(args ...interface{}) {
	a.log(SeverityAlert, fmt.Sprint(args...))
}

func (a *AppLogger) Alertf(format string, args ...interface{}) {
	a.log(SeverityAlert, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Alertln(args ...interface{}) {
	a.log(SeverityAlert, fmt.Sprintln(args...))
}

func (a *AppLogger) Emergency(args ...interface{}) {
	a.log(SeverityEmergency, fmt.Sprint(args...))
}

func (a *AppLogger) Emergencyf(format string, args ...interface{}) {
	a.log(SeverityEmergency, fmt.Sprintf(format, args...))
}

func (a *AppLogger) Emergencyln(args ...interface{}) {
	a.log(SeverityEmergency, fmt.Sprintln(args...))
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
		"logType":                      "app_log",
	}
	b, err := json.Marshal(appLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
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
