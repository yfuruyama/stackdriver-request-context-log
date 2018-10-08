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
		// appLogger := stackdriver.RequestContextLogger(r) // logrus.Logger
		// appLogger.Infof("hello world")
		fmt.Fprintf(w, "OK\n")
	})

	logger := stackdriver.NewLogger(os.Stderr, os.Stdout)
	handler := stackdriver.Handler(logger, mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
