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

func WithOut(outRequestLog io.Writer, outContextLog io.Writer) Option {
	return func(c *Config) {
		c.requestLogger = log.New(outRequestLog, "", 0)
		c.contextLogger = log.New(outContextLog, "", 0)
	}
}

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

type ContextLogger struct {
	logger         *log.Logger
	Trace          string
	Severity       Severity
	loggedSeverity []Severity
}

func RequestContextLogger(r *http.Request) *ContextLogger {
	v := r.Context().Value(contextLoggerKey)
	if l, ok := v.(*ContextLogger); ok {
		return l
	} else {
		return nil
	}
}

func (l *ContextLogger) Default(args ...interface{}) {
	l.log(SeverityDefault, fmt.Sprint(args...))
}

func (l *ContextLogger) Defaultf(format string, args ...interface{}) {
	l.log(SeverityDefault, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Defaultln(args ...interface{}) {
	l.log(SeverityDefault, fmt.Sprintln(args...))
}

func (l *ContextLogger) Debug(args ...interface{}) {
	l.log(SeverityDebug, fmt.Sprint(args...))
}

func (l *ContextLogger) Debugf(format string, args ...interface{}) {
	l.log(SeverityDebug, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Debugln(args ...interface{}) {
	l.log(SeverityDebug, fmt.Sprintln(args...))
}

func (l *ContextLogger) Info(args ...interface{}) {
	l.log(SeverityInfo, fmt.Sprint(args...))
}

func (l *ContextLogger) Infof(format string, args ...interface{}) {
	l.log(SeverityInfo, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Infoln(args ...interface{}) {
	l.log(SeverityInfo, fmt.Sprintln(args...))
}

func (l *ContextLogger) Notice(args ...interface{}) {
	l.log(SeverityNotice, fmt.Sprint(args...))
}

func (l *ContextLogger) Noticef(format string, args ...interface{}) {
	l.log(SeverityNotice, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Noticeln(args ...interface{}) {
	l.log(SeverityNotice, fmt.Sprintln(args...))
}

func (l *ContextLogger) Warning(args ...interface{}) {
	l.log(SeverityWarning, fmt.Sprint(args...))
}

func (l *ContextLogger) Warningf(format string, args ...interface{}) {
	l.log(SeverityWarning, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Warningln(args ...interface{}) {
	l.log(SeverityWarning, fmt.Sprintln(args...))
}

func (l *ContextLogger) Warn(args ...interface{}) {
	l.Warning(args...)
}

func (l *ContextLogger) Warnf(format string, args ...interface{}) {
	l.Warningf(format, args...)
}

func (l *ContextLogger) Warnln(args ...interface{}) {
	l.Warningln(args...)
}

func (l *ContextLogger) Error(args ...interface{}) {
	l.log(SeverityError, fmt.Sprint(args...))
}

func (l *ContextLogger) Errorf(format string, args ...interface{}) {
	l.log(SeverityError, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Errorln(args ...interface{}) {
	l.log(SeverityError, fmt.Sprintln(args...))
}

func (l *ContextLogger) Critical(args ...interface{}) {
	l.log(SeverityCritical, fmt.Sprint(args...))
}

func (l *ContextLogger) Criticalf(format string, args ...interface{}) {
	l.log(SeverityCritical, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Criticalln(args ...interface{}) {
	l.log(SeverityCritical, fmt.Sprintln(args...))
}

func (l *ContextLogger) Alert(args ...interface{}) {
	l.log(SeverityAlert, fmt.Sprint(args...))
}

func (l *ContextLogger) Alertf(format string, args ...interface{}) {
	l.log(SeverityAlert, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Alertln(args ...interface{}) {
	l.log(SeverityAlert, fmt.Sprintln(args...))
}

func (l *ContextLogger) Emergency(args ...interface{}) {
	l.log(SeverityEmergency, fmt.Sprint(args...))
}

func (l *ContextLogger) Emergencyf(format string, args ...interface{}) {
	l.log(SeverityEmergency, fmt.Sprintf(format, args...))
}

func (l *ContextLogger) Emergencyln(args ...interface{}) {
	l.log(SeverityEmergency, fmt.Sprintln(args...))
}

func (l *ContextLogger) log(severity Severity, msg string) {
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
