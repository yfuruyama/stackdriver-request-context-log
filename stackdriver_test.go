package stackdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestIntegration(t *testing.T) {
	r, _ := http.NewRequest("GET", "/foo?bar=baz", nil)
	r.Header.Add("User-Agent", "test")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		logger := RequestContextLogger(r)
		logger.Debugf("1")
		logger.Infof("2")
		logger.Warnf("3")
		logger.Errorf("4")

		fmt.Fprintf(w, "OK\n")
	})

	requestLogOut := new(bytes.Buffer)
	contextLogOut := new(bytes.Buffer)

	config := NewConfig("test")
	config.RequestLogOut = requestLogOut
	config.ContextLogOut = contextLogOut
	config.Severity = SeverityInfo
	config.AdditionalData = AdditionalData{
		"service": "foo",
		"version": 1.0,
	}
	handler := RequestLogging(config)(mux)
	handler.ServeHTTP(w, r)

	// check request log
	var httpRequestLog HttpRequestLog
	err := json.Unmarshal(requestLogOut.Bytes(), &httpRequestLog)
	if err != nil {
		t.Fatal(err)
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(HttpRequestLog{}, "Time", "Trace"),
		cmpopts.IgnoreFields(HttpRequest{}, "RemoteIp", "ServerIp", "Latency"),
	}
	expected := HttpRequestLog{
		Severity: "ERROR",
		AdditionalData: AdditionalData{
			"service": "foo",
			"version": 1.0,
		},
		HttpRequest: HttpRequest{
			RequestMethod:                  "GET",
			RequestUrl:                     "/foo?bar=baz",
			RequestSize:                    "0",
			Status:                         200,
			ResponseSize:                   "3",
			UserAgent:                      "test",
			Referer:                        "",
			CacheLookup:                    false,
			CacheHit:                       false,
			CacheValidatedWithOriginServer: false,
			Protocol:                       "HTTP/1.1",
		},
	}

	if !cmp.Equal(httpRequestLog, expected, opts...) {
		t.Errorf("diff: %s", cmp.Diff(httpRequestLog, expected, opts...))
	}

	if !strings.HasSuffix(httpRequestLog.HttpRequest.Latency, "s") {
		t.Errorf("invalid latency: %s", httpRequestLog.HttpRequest.Latency)
	}

	// check context log
	logs := strings.Split(string(contextLogOut.Bytes()), "\n")
	logs = logs[:len(logs)-1]
	logExpected := []struct {
		Severity string
		Message  string
	}{
		{"INFO", "2"},
		{"WARNING", "3"},
		{"ERROR", "4"},
	}
	for idx, log := range logs {
		var cLog contextLog
		if err := json.Unmarshal([]byte(log), &cLog); err != nil {
			t.Fatal(err)
		}
		expected := contextLog{
			Severity: logExpected[idx].Severity,
			Message:  logExpected[idx].Message,
			AdditionalData: AdditionalData{
				"service": "foo",
				"version": 1.0,
			},
		}
		opts := []cmp.Option{
			cmpopts.IgnoreFields(contextLog{}, "Time", "Trace", "SourceLocation"),
		}
		if !cmp.Equal(cLog, expected, opts...) {
			t.Errorf("diff: %s", cmp.Diff(cLog, expected, opts...))
		}

		// check trace
		if httpRequestLog.Trace != cLog.Trace {
			t.Errorf("different trace: httpRequestLog=%s, contextLog=%s", httpRequestLog.Trace, cLog.Trace)
		}
	}
}

func TestNoContextLog(t *testing.T) {
	r, _ := http.NewRequest("GET", "/foo?bar=baz", nil)
	r.Header.Add("User-Agent", "test")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
	})

	requestLogOut := new(bytes.Buffer)
	contextLogOut := new(bytes.Buffer)

	config := NewConfig("test")
	config.RequestLogOut = requestLogOut
	config.ContextLogOut = contextLogOut
	handler := RequestLogging(config)(mux)
	handler.ServeHTTP(w, r)

	// check request log
	var httpRequestLog HttpRequestLog
	err := json.Unmarshal(requestLogOut.Bytes(), &httpRequestLog)
	if err != nil {
		t.Fatal(err)
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(HttpRequestLog{}, "Time", "Trace"),
		cmpopts.IgnoreFields(HttpRequest{}, "RemoteIp", "ServerIp", "Latency"),
	}
	expected := HttpRequestLog{
		Severity:       "DEFAULT",
		AdditionalData: nil,
		HttpRequest: HttpRequest{
			RequestMethod:                  "GET",
			RequestUrl:                     "/foo?bar=baz",
			RequestSize:                    "0",
			Status:                         200,
			ResponseSize:                   "3",
			UserAgent:                      "test",
			Referer:                        "",
			CacheLookup:                    false,
			CacheHit:                       false,
			CacheValidatedWithOriginServer: false,
			Protocol:                       "HTTP/1.1",
		},
	}

	if !cmp.Equal(httpRequestLog, expected, opts...) {
		t.Errorf("diff: %s", cmp.Diff(httpRequestLog, expected, opts...))
	}

	if len(contextLogOut.Bytes()) != 0 {
		t.Errorf("context log exists: %s", string(contextLogOut.Bytes()))
	}
}
