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
	config := log.NewConfig(projectId,
		log.WithRequestLogOut(os.Stderr),   // set output for request log
		log.WithContextLogOut(os.Stdout),   // set output for context log
		log.WithSeverity(log.SeverityInfo), // set severity
		log.WithAdditionalFields(log.AdditionalFields{ // set additional fields for request logging
			"service": "foo",
			"version": 1.0,
		}),
	)
	router.Use(log.RequestLogging(config))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		logger := log.RequestContextLogger(r)

		logger.Debugf("Hi")
		logger.Infof("Hello")
		logger.Warnf("World")
		fmt.Fprintf(w, "OK\n")
	})

	fmt.Println("Waiting requests on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}