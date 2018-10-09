stackdriver-request-context-log
===

Stackdriver Logging Go library for grouping a request log and application logs.

With this library, a request log is automatically logged and every application logs within the request are grouped together (similtar to App Engine).

<img alt="screenshot" src="https://github.com/yfuruyama/stackdriver-request-context-log/blob/master/img/screenshot.png">

Note that the interface of this library is still **ALPHA** level quality.  
Breaking changes will be introduced frequently.

## Install

```
go get -u github.com/yfuruyama/stackdriver-request-context-log
```

## Example

This simple example shows how to integrate this library into your web application.

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/yfuruyama/stackdriver-request-context-log"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := log.RequestContextLogger(r) // thig logger's log is grouped with request log
		logger.Debugf("Hi")
		logger.Infof("Hello")
		logger.Warnf("World")

		fmt.Fprintf(w, "OK\n")
	})

	projectId := "my-gcp-project"
	config := log.NewConfig(projectId)
	config.RequestLogOut = os.Stderr                // set output for request log
	config.ContextLogOut = os.Stdout                // set output for context log
	config.Severity = log.SeverityInfo              // set severity
	config.AdditionalFields = log.AdditionalFields{ // set additional fields for request logging
		"service": "foo",
		"version": 1.0,
	}

	handler := log.RequestLogging(config)(mux)

	fmt.Println("Waiting requests on port 8010...")
	if err := http.ListenAndServe(":8010", handler); err != nil {
		panic(err)
	}
}
```

When this application receives a HTTP request `GET /`, following logs will be logged (with pretty print for display purposes only).

```json
// STDOUT
{
  "logType": "context_log",
  "logging.googleapis.com/trace": "projects/my-gcp-project/traces/5e328d9926f7bb7bb15fdbafa5b08439",
  "message": "Hello",
  "severity": "INFO",
  "time": "2018-10-09T18:21:43.629731+09:00"
}
{
  "logType": "context_log",
  "logging.googleapis.com/trace": "projects/my-gcp-project/traces/5e328d9926f7bb7bb15fdbafa5b08439",
  "message": "World",
  "severity": "WARNING",
  "time": "2018-10-09T18:21:43.63184+09:00"
}

// STDERR
{
  "httpRequest": {
    "cacheHit": false,
    "cacheLookUp": false,
    "cacheValidatedWithOriginServer": false,
    "latency": "0.007073s",
    "protocol": "HTTP/1.1",
    "referer": "",
    "remoteIp": "[::1]:61502",
    "requestMethod": "GET",
    "requestSize": "0",
    "requestUrl": "/",
    "responseSize": "3",
    "serverIp": "",
    "status": 200,
    "userAgent": "curl/7.58.0"
  },
  "logType": "request_log",
  "logging.googleapis.com/trace": "projects/my-gcp-project/traces/5e328d9926f7bb7bb15fdbafa5b08439",
  "service": "foo",
  "severity": "WARNING",
  "time": "2018-10-09T18:21:43.632127+09:00",
  "version": 1
}
```

The log format is based on [LogEntry](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry)'s structured payload so that you can pass these logs to [Stackdriver Logging agent](https://cloud.google.com/logging/docs/agent/).  

Some of fields are treated specially at logging agent. See more details: https://cloud.google.com/logging/docs/agent/configuration?hl=en#special_fields_in_structured_payloads

## How logs are grouped

This library leverages the grouping feature of Stackdriver Logging.
See following references fore more details. 

* https://godoc.org/cloud.google.com/go/logging#hdr-Grouping_Logs_by_Request
* https://cloud.google.com/appengine/articles/logging#linking_app_logs_and_requests
