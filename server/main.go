package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/adrone13/go-long-poll/pkg/logging"
	"github.com/adrone13/go-long-poll/server/handler"
	"github.com/adrone13/go-long-poll/server/jobs"
)

var (
	jobSteps     = 10
	jobStepDelay = time.Second
)

func main() {
	jobStore := jobs.New()
	logger := logging.New("debug", true)

	jobsHandler := handler.NewJobsHandler(logger, jobStore, jobSteps, jobStepDelay)

	http.HandleFunc("/submit", jobsHandler.SubmitHandler)
	http.HandleFunc("/result", jobsHandler.ResultHandler)

	logger.Info("starting server", slog.String("addr", ":8080"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
