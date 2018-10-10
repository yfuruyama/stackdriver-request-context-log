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
		logger := log.RequestContextLogger(r) // context logger's logs are grouped with request log
		logger.Debugf("Hi")
		logger.Infof("Hello")
		logger.Warnf("World")

		fmt.Fprintf(w, "OK\n")
	})

	projectId := "my-gcp-project"
	config := log.NewConfig(projectId)
	config.RequestLogOut = os.Stderr                // request log to stderr
	config.ContextLogOut = os.Stdout                // context log to stdout
	config.Severity = log.SeverityInfo              // only over INFO logs are logged
	config.AdditionalFields = log.AdditionalFields{ // set additional fields for all logs
		"service": "foo",
		"version": 1.0,
	}

	handler := log.RequestLogging(config)(mux)

	fmt.Println("Waiting requests on port 8080...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
