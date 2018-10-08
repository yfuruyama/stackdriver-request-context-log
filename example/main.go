package main

import (
	"fmt"
	"net/http"
	"os"

	stackdriver "github.com/yfuruyama/stackdriver-request-context-log"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		appLogger := stackdriver.LoggerFromRequest(r) // logrus.Logger
		appLogger.Infof("This is an info log")
		appLogger.Warnf("This is a warning log")
		fmt.Fprintf(w, "OK\n")
	})

	projectId := "yfuruyama-sandbox"
	logger := stackdriver.NewLogger(os.Stderr, os.Stdout, projectId)
	handler := stackdriver.Handler(logger, mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
