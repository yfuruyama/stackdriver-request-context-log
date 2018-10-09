package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	stackdriver "github.com/yfuruyama/stackdriver-request-context-log"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := stackdriver.RequestContextLogger(r)
		logger.Debugf("This is a debug log")
		logger.Infof("This is an info log")
		logger.Warnf("This is a warning log")
		logger.Errorf("This is an error log")
		logger.Alertf("This is an alert log")
		logger.Emergency("This is an emergency log")
		fmt.Fprintf(w, "OK\n")
	})

	projectId, _ := getDefaultProjectId()
	config := stackdriver.NewConfig(projectId,
		stackdriver.WithRequestLogOut(os.Stderr),
		stackdriver.WithContextLogOut(os.Stdout),
		stackdriver.WithSeverity(stackdriver.SeverityInfo),
		stackdriver.WithAdditionalFields(stackdriver.AdditionalFields{
			"service": "foo",
			"version": 1.0,
		}),
	)
	handler := stackdriver.Handler(config, mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}

func getDefaultProjectId() (string, error) {
	out, err := exec.Command("gcloud", "config", "list", "--format", "value(core.project)").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
