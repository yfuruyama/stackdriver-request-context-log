stackdriver-request-context-log
===

Stackdriver Logging Go library for grouping a request log and application logs.

With this library, a request log is automatically logged and each application logs within the request are grouped each together, like App Engine logs.

[put image here]

Note that this library's interface is **ALPHA** level quality.  
Breaking changes will be introduced frequently.

# Install

```
go get -u github.com/yfuruyama/stackdriver-request-context-log
```

# Example

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
	config := log.NewConfig(projectId,
		log.WithRequestLogOut(os.Stderr),   // set output for request log
		log.WithContextLogOut(os.Stdout),   // set output for context log
		log.WithSeverity(log.SeverityInfo), // set severity
		log.WithAdditionalFields(log.AdditionalFields{ // set additional fields for request logging
			"service": "foo",
			"version": 1.0,
		}),
	)
	handler := log.RequestLogging(config)(mux)

	fmt.Println("Waiting requests on port 8080...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
```

When this programs receives HTTP request "GET /", logging library outputs following logs which conforms Stackdriver Logging format.  
You can pass these logs to gcp-fluentd to send them to Stackdriver Logging.

```json
// STDOUT
{"logType":"context_log","logging.googleapis.com/trace":"projects/my-gcp-project/traces/5e328d9926f7bb7bb15fdbafa5b08439","message":"Hello","severity":"INFO","time":"2018-10-09T18:21:43.629731+09:00"}
{"logType":"context_log","logging.googleapis.com/trace":"projects/my-gcp-project/traces/5e328d9926f7bb7bb15fdbafa5b08439","message":"World","severity":"WARNING","time":"2018-10-09T18:21:43.63184+09:00"}

// STDERR
{"httpRequest":{"cacheHit":false,"cacheLookUp":false,"cacheValidatedWithOriginServer":false,"latency":"0.007073s","protocol":"HTTP/1.1","referer":"","remoteIp":"[::1]:61502","requestMethod":"GET","requestSize":"0","requestUrl":"/","responseSize":"3","serverIp":"","status":200,"userAgent":"curl/7.58.0"},"logType":"request_log","logging.googleapis.com/trace":"projects/my-gcp-project/traces/5e328d9926f7bb7bb15fdbafa5b08439","service":"foo","severity":"WARNING","time":"2018-10-09T18:21:43.632127+09:00","version":1}
```

TODO
===

* test
* GoDoc
* Add more example
* interface redesign
* traceId generation
