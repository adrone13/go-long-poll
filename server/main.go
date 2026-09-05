package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adrone13/go-long-poll/pkg/logging"
)

type Jobs struct {
	items map[string]chan string
	m     sync.RWMutex
}

var (
	jobs = Jobs{
		items: make(map[string]chan string),
	}
	logger   = logging.New("debug", true)
	jobIDSeq atomic.Int64
)

var (
	jobSteps     = 10
	jobStepDelay = time.Second
)

func performJob(jobID string) {
	logger.Info("job started", slog.String("job_id", jobID))

	for i := range jobSteps {
		<-time.After(jobStepDelay)
		logger.Info(fmt.Sprintf("job progress: %d%%", (i+1)*10), slog.String("job_id", jobID))
	}

	jobs.m.RLock()
	resCh := jobs.items[jobID]
	jobs.m.RUnlock()

	resCh <- fmt.Sprintf("job %s is done", jobID)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("/%s", r.URL.Path[1:]))

	jobID := strconv.FormatInt(jobIDSeq.Add(1), 10)

	jobs.m.Lock()
	jobs.items[jobID] = make(chan string, 1)
	jobs.m.Unlock()
	go performJob(jobID)

	logger.Info("job created", slog.String("job_id", jobID))

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(jobID)); err != nil {
		logger.Error("failed to job id", slog.String("error", err.Error()))
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	logger.Info(fmt.Sprintf("/%s", r.URL.Path[1:]))

	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		logger.Error("failed to get result", slog.String("error", "job_id is required"))
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	jobs.m.RLock()
	resCh, ok := jobs.items[jobID]
	jobs.m.RUnlock()
	if !ok {
		logger.Error("job not found", slog.String("job_id", jobID))
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	select {
	case res := <-resCh:
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(res)); err != nil {
			logger.Error("failed to write result", slog.String("error", err.Error()))
		}

		jobs.m.Lock()
		delete(jobs.items, jobID)
		jobs.m.Unlock()

	case <-ctx.Done():
		w.WriteHeader(http.StatusRequestTimeout)
	}
}

func main() {
	logger.Info("starting server", slog.String("addr", ":8080"))
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/result", resultHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
