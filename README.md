stackdriver-request-context-log
===
[![CircleCI](https://circleci.com/gh/yfuruyama/stackdriver-request-context-log.svg?style=svg)](https://circleci.com/gh/yfuruyama/stackdriver-request-context-log)

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

	// Set request handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get request context logger
		logger := log.RequestContextLogger(r)

		// These logs are grouped with the request log
		logger.Debugf("Hi")
		logger.Infof("Hello")
		logger.Warnf("World")

		fmt.Fprintf(w, "OK\n")
	})

	projectId := "my-gcp-project"

	// Make config for this library
	config := log.NewConfig(projectId)
	config.RequestLogOut = os.Stderr            // request log to stderr
	config.ContextLogOut = os.Stdout            // context log to stdout
	config.Severity = log.SeverityInfo          // only over INFO logs are logged
	config.AdditionalData = log.AdditionalData{ // set additional fields for all logs
		"service": "foo",
		"version": 1.0,
	}

	// Set middleware for the request log to be automatically logged
	handler := log.RequestLogging(config)(mux)

	// Run server
	fmt.Println("Waiting requests on port 8080...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
```

When this application receives a HTTP request `GET /`, following logs will be logged (with pretty print for display purposes only).

```json
// STDOUT
{
  "time": "2018-10-10T16:46:07.476567+09:00",
  "logging.googleapis.com/trace": "projects/my-gcp-project/traces/a8cb3e640add456cf7ed58e4a0589ea0",
  "logging.googleapis.com/sourceLocation": {
    "file": "main.go",
    "line": "21",
    "function": "main.main.func1"
  },
  "severity": "INFO",
  "message": "Hello",
  "data": {
    "service": "foo",
    "version": 1
  }
}
{
  "time": "2018-10-10T16:46:07.476806+09:00",
  "logging.googleapis.com/trace": "projects/my-gcp-project/traces/a8cb3e640add456cf7ed58e4a0589ea0",
  "logging.googleapis.com/sourceLocation": {
    "file": "main.go",
    "line": "22",
    "function": "main.main.func1"
  },
  "severity": "WARNING",
  "message": "World",
  "data": {
    "service": "foo",
    "version": 1
  }
}

// STDERR
{
  "time": "2018-10-10T16:46:07.47682+09:00",
  "logging.googleapis.com/trace": "projects/my-gcp-project/traces/a8cb3e640add456cf7ed58e4a0589ea0",
  "severity": "WARNING",
  "httpRequest": {
    "requestMethod": "GET",
    "requestUrl": "/",
    "requestSize": "0",
    "status": 200,
    "responseSize": "3",
    "userAgent": "curl/7.58.0",
    "remoteIp": "[::1]:61352",
    "serverIp": "192.168.86.31",
    "referer": "",
    "latency": "0.000304s",
    "cacheLookup": false,
    "cacheHit": false,
    "cacheValidatedWithOriginServer": false,
    "protocol": "HTTP/1.1"
  },
  "data": {
    "service": "foo",
    "version": 1
  }
}
```

The log format is based on [LogEntry](https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry)'s structured payload so that you can pass these logs to [Stackdriver Logging agent](https://cloud.google.com/logging/docs/agent/).  

## How logs are grouped

This library leverages the grouping feature of Stackdriver Logging.
See following references fore more details. 

* https://godoc.org/cloud.google.com/go/logging#hdr-Grouping_Logs_by_Request
* https://cloud.google.com/appengine/articles/logging#linking_app_logs_and_requests
