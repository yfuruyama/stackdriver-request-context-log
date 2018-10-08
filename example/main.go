package main

import (
	"net/http"
	"os"
)

func main() {
	// httpLogger := logrus.New()
	// httpLogger.Out = os.Stderr
	// httpLogger.Format = stackdriver.NewHttpRequestFormatter()
	// httpLogger.Level = logrus.InfoLevel

	// appLogger := logrus.New()
	// appLogger.Out = os.Stdout
	// appLogger.Format = stackdriver.NewAppLogFormatter()
	// appLogger.Level = logrus.InfoLevel

	// router := chi.NewRouter()
	// router.Use(stackdriver.NewMiddleware(httpLogger, appLogger))
	// router.Get(func(w http.ResponseWriter, r *http.Request) {
	// appLogger := stackdriver.GetAppLogger(r)
	// appLogger.Infof("hello world")
	// appLogger.Warnf("hello world")
	// })

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		appLogger := stackdriver.GetAppLogger(r) // logrus.Logger
		appLogger.Infof("hello world")
	})

	logger := stackdriver.NewLogger(os.Stderr, os.Stdout, stackdriver.InfoLevel)
	handler = stackdriver.Handler(logger, mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
