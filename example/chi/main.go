package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/go-chi/chi"
	stackdriver "github.com/yfuruyama/stackdriver-request-context-log"
)

func main() {
	router := chi.NewRouter()

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
	router.Use(stackdriver.RequestLogging(config))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		logger := stackdriver.RequestContextLogger(r)
		logger.Debugf("This is a debug log")
		logger.Infof("This is an info log")
		logger.Errorf("This is an error log")
		fmt.Fprintf(w, "OK\n")
	})

	fmt.Println("Waiting a request on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
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
