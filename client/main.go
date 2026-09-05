package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/adrone13/go-long-poll/pkg/logging"
)

func main() {
	logger := logging.New("debug", true)

	submitRes, err := http.Get("http://localhost:8080/submit")
	if err != nil {
		logger.Error("failed to submit job", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer submitRes.Body.Close()

	body, err := io.ReadAll(submitRes.Body)
	if err != nil {
		logger.Error("failed to read body", slog.String("error", err.Error()))
		os.Exit(1)
	}
	jobID := string(body)

	resultRes, err := http.Get(fmt.Sprintf("http://localhost:8080/result?job_id=%s", jobID))
	if err != nil {
		logger.Error("failed to get result", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer resultRes.Body.Close()

	body, err = io.ReadAll(resultRes.Body)
	if err != nil {
		logger.Error("failed to read body", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info(string(body))
}
