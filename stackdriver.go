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

type Config struct {
	requestLogger *log.Logger
	contextLogger *log.Logger
	projectId     string
	additional    AdditionalFields
	severity      Severity
}

type Option func(*Config)

func WithSeverity(severity Severity) Option {
	return func(c *Config) {
		c.severity = severity
	}
}

func WithAdditionalFields(fields AdditionalFields) Option {
	return func(c *Config) {
		c.additional = fields
	}
}

func WithRequestLogOut(out io.Writer) Option {
	return func(c *Config) {
		c.requestLogger = log.New(out, "", 0)
	}
}

func WithContextLogOut(out io.Writer) Option {
	return func(c *Config) {
		c.contextLogger = log.New(out, "", 0)
	}
}

// NewConfig creates a config which is passed to `Handler` function.
func NewConfig(projectId string, options ...Option) *Config {
	config := &Config{
		projectId:     projectId,
		severity:      SeverityInfo,
		requestLogger: log.New(os.Stderr, "", 0),
		contextLogger: log.New(os.Stdout, "", 0),
	}
	for _, option := range options {
		option(config)
	}
	return config
}

// Severity is the level of log. More details:
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

// String returns text representation for the severity
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

// ContextLogger is the logger which is combined with the request
type ContextLogger struct {
	logger         *log.Logger
	Trace          string
	Severity       Severity
	loggedSeverity []Severity
}

// RequestContextLogger gets request-context logger for the request
func RequestContextLogger(r *http.Request) *ContextLogger {
	v := r.Context().Value(contextLoggerKey)
	if l, ok := v.(*ContextLogger); ok {
		return l
	} else {
		return nil
	}
}

// Default logs a message at DEFAULT severity
func (l *ContextLogger) Default(args ...interface{}) {
	l.write(SeverityDefault, fmt.Sprint(args...))
}

// Defaultf logs a message at DEFAULT severity
func (l *ContextLogger) Defaultf(format string, args ...interface{}) {
	l.write(SeverityDefault, fmt.Sprintf(format, args...))
}

// Defaultln logs a message at DEFAULT severity
func (l *ContextLogger) Defaultln(args ...interface{}) {
	l.write(SeverityDefault, fmt.Sprintln(args...))
}

// Debug logs a message at DEBUG severity
func (l *ContextLogger) Debug(args ...interface{}) {
	l.write(SeverityDebug, fmt.Sprint(args...))
}

// Debugf logs a message at DEBUG severity
func (l *ContextLogger) Debugf(format string, args ...interface{}) {
	l.write(SeverityDebug, fmt.Sprintf(format, args...))
}

// Debugln logs a message at DEBUG severity
func (l *ContextLogger) Debugln(args ...interface{}) {
	l.write(SeverityDebug, fmt.Sprintln(args...))
}

// Info logs a message at INFO severity
func (l *ContextLogger) Info(args ...interface{}) {
	l.write(SeverityInfo, fmt.Sprint(args...))
}

// Infof logs a message at INFO severity
func (l *ContextLogger) Infof(format string, args ...interface{}) {
	l.write(SeverityInfo, fmt.Sprintf(format, args...))
}

// Infofln logs a message at INFO severity
func (l *ContextLogger) Infoln(args ...interface{}) {
	l.write(SeverityInfo, fmt.Sprintln(args...))
}

// Notice logs a message at NOTICE severity
func (l *ContextLogger) Notice(args ...interface{}) {
	l.write(SeverityNotice, fmt.Sprint(args...))
}

// Noticef logs a message at NOTICE severity
func (l *ContextLogger) Noticef(format string, args ...interface{}) {
	l.write(SeverityNotice, fmt.Sprintf(format, args...))
}

// Noticeln logs a message at NOTICE severity
func (l *ContextLogger) Noticeln(args ...interface{}) {
	l.write(SeverityNotice, fmt.Sprintln(args...))
}

// Warning logs a message at WARNING severity
func (l *ContextLogger) Warning(args ...interface{}) {
	l.write(SeverityWarning, fmt.Sprint(args...))
}

// Warningf logs a message at WARNING severity
func (l *ContextLogger) Warningf(format string, args ...interface{}) {
	l.write(SeverityWarning, fmt.Sprintf(format, args...))
}

// Warningln logs a message at WARNING severity
func (l *ContextLogger) Warningln(args ...interface{}) {
	l.write(SeverityWarning, fmt.Sprintln(args...))
}

// Warn is alias for Warning
func (l *ContextLogger) Warn(args ...interface{}) {
	l.Warning(args...)
}

// Warnf is alias for Warningf
func (l *ContextLogger) Warnf(format string, args ...interface{}) {
	l.Warningf(format, args...)
}

// Warnln is alias for Warningln
func (l *ContextLogger) Warnln(args ...interface{}) {
	l.Warningln(args...)
}

// Error logs a message at ERROR severity
func (l *ContextLogger) Error(args ...interface{}) {
	l.write(SeverityError, fmt.Sprint(args...))
}

// Errorf logs a message at ERROR severity
func (l *ContextLogger) Errorf(format string, args ...interface{}) {
	l.write(SeverityError, fmt.Sprintf(format, args...))
}

// Errorln logs a message at ERROR severity
func (l *ContextLogger) Errorln(args ...interface{}) {
	l.write(SeverityError, fmt.Sprintln(args...))
}

// Critical logs a message at CRITICAL severity
func (l *ContextLogger) Critical(args ...interface{}) {
	l.write(SeverityCritical, fmt.Sprint(args...))
}

// Criticalf logs a message at CRITICAL severity
func (l *ContextLogger) Criticalf(format string, args ...interface{}) {
	l.write(SeverityCritical, fmt.Sprintf(format, args...))
}

// Criticalln logs a message at CRITICAL severity
func (l *ContextLogger) Criticalln(args ...interface{}) {
	l.write(SeverityCritical, fmt.Sprintln(args...))
}

// Alert logs a message at ALERT severity
func (l *ContextLogger) Alert(args ...interface{}) {
	l.write(SeverityAlert, fmt.Sprint(args...))
}

// Alertf logs a message at ALERT severity
func (l *ContextLogger) Alertf(format string, args ...interface{}) {
	l.write(SeverityAlert, fmt.Sprintf(format, args...))
}

// Alertln logs a message at ALERT severity
func (l *ContextLogger) Alertln(args ...interface{}) {
	l.write(SeverityAlert, fmt.Sprintln(args...))
}

// Emergency logs a message at EMERGENCY severity
func (l *ContextLogger) Emergency(args ...interface{}) {
	l.write(SeverityEmergency, fmt.Sprint(args...))
}

// Emergencyf logs a message at EMERGENCY severity
func (l *ContextLogger) Emergencyf(format string, args ...interface{}) {
	l.write(SeverityEmergency, fmt.Sprintf(format, args...))
}

// Emergencyln logs a message at EMERGENCY severity
func (l *ContextLogger) Emergencyln(args ...interface{}) {
	l.write(SeverityEmergency, fmt.Sprintln(args...))
}

func (l *ContextLogger) write(severity Severity, msg string) {
	if severity < l.Severity {
		return
	}

	l.loggedSeverity = append(l.loggedSeverity, severity)

	contextLog := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339Nano),
		"logging.googleapis.com/trace": l.Trace,
		"severity":                     severity.String(),
		"message":                      msg,
		"logType":                      "context_log",
	}
	b, err := json.Marshal(contextLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	l.logger.Println(string(b))
}

func (l *ContextLogger) maxSeverity() Severity {
	max := SeverityDefault
	for _, s := range l.loggedSeverity {
		if s > max {
			max = s
		}
	}
	return max
}
