package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	log "github.com/yfuruyama/stackdriver-request-context-log"
)

func main() {
	router := chi.NewRouter()

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
	router.Use(log.RequestLogging(config))

	// Set request handler
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Get request context logger
		logger := log.RequestContextLogger(r)

		// These logs are grouped with the request log
		logger.Debugf("Hi")
		logger.Infof("Hello")
		logger.Warnf("World")

		fmt.Fprintf(w, "OK\n")
	})

	// Run server
	fmt.Println("Waiting requests on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
