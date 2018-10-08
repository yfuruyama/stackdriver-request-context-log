package stackdriver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	requestLogger *logrus.Logger
	appLogger     *logrus.Logger
	projectId     string
	// Level: int
}

func NewLogger(outRequestLog io.Writer, outAppLog io.Writer, projectId string) *Logger {
	requestLogger := logrus.New()
	requestLogger.Out = outRequestLog
	// requestLogger.Formatter = &logrus.JSONFormatter{
	// DisableColors:    true,
	// DisableTimestamp: true,
	// }

	appLogger := logrus.New()
	appLogger.Out = outAppLog

	return &Logger{
		requestLogger: requestLogger,
		appLogger:     appLogger,
		projectId:     projectId,
	}
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
	CacheLookUp                    bool   `json:"cacheLookUp"`
	CacheHit                       bool   `json:"cacheHit"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer"`
	Protocol                       string `json:"protocol"`
}

type HttpRequestLog struct {
	Time string `json:"time"`
	// Message string `json:"message"`
	Trace       string       `json:"logging.googleapis.com/trace"`
	Severity    string       `json:"severity"`
	HttpRequest *HttpRequest `json:"httpRequest"`
	// LogType string `json:"logType"`
}

func (l *Logger) WriteRequestLog(r *http.Request, status int, responseSize int, elapsed time.Duration) error {
	traceId := r.Context().Value("traceId").(string)
	trace := fmt.Sprintf("projects/%s/traces/%s", l.projectId, traceId)

	latency := fmt.Sprintf("%fs", elapsed.Seconds())

	requestLog := &HttpRequestLog{
		Time:     time.Now().Format(time.RFC3339Nano),
		Trace:    trace,
		Severity: "INFO", // TODO
		HttpRequest: &HttpRequest{
			RequestMethod:                  r.Method,
			RequestUrl:                     r.URL.Path,
			RequestSize:                    fmt.Sprintf("%d", r.ContentLength),
			Status:                         status,
			ResponseSize:                   fmt.Sprintf("%d", responseSize),
			UserAgent:                      r.UserAgent(),
			RemoteIp:                       r.RemoteAddr,
			ServerIp:                       "localhost",
			Referer:                        r.Referer(),
			Latency:                        latency,
			CacheLookUp:                    false, // TODO
			CacheHit:                       false, // TODO
			CacheValidatedWithOriginServer: false, // TODO
			Protocol:                       r.Proto,
		},
		// "logType": "request_log",
	}
	requestLogJson, err := json.Marshal(requestLog)
	if err != nil {
		return err
	}

	l.requestLogger.Println(string(requestLogJson))

	return nil
}

// func RequestContextLogger(r *http.Request) logrus.Logger {
// }
