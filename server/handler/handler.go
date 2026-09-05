package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/adrone13/go-long-poll/server/jobs"
)

type JobHandler struct {
	logger       *slog.Logger
	jobStore     *jobs.Store
	jobSteps     int
	jobStepDelay time.Duration
}

func NewJobsHandler(logger *slog.Logger, jobStore *jobs.Store, jobSteps int, jobStepsDelay time.Duration) *JobHandler {
	return &JobHandler{
		logger:       logger,
		jobStore:     jobStore,
		jobSteps:     jobSteps,
		jobStepDelay: jobStepsDelay,
	}
}

func (jh *JobHandler) performJob(jobID jobs.JobID) {
	jh.logger.Info("job started", slog.Any("job_id", jobID))

	for i := range jh.jobSteps {
		<-time.After(jh.jobStepDelay)
		jh.logger.Info(fmt.Sprintf("job progress: %d%%", (i+1)*10), slog.Any("job_id", jobID))
	}

	ch, ok := jh.jobStore.Get(jobID)
	if !ok {
		jh.logger.Error("perform job: not found")
		return
	}

	ch <- fmt.Sprintf("job %d is done", jobID)
}

func (jh *JobHandler) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	jh.logger.Info(fmt.Sprintf("/%s", r.URL.Path[1:]))

	jobID := jh.jobStore.Add()

	go jh.performJob(jobID)

	jh.logger.Info("job created", slog.String("job_id", jobID.String()))

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(jobID.String())); err != nil {
		jh.logger.Error("failed to job id", slog.String("error", err.Error()))
	}
}

func (jh *JobHandler) ResultHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	jh.logger.Info(fmt.Sprintf("/%s", r.URL.Path[1:]))

	jobIDStr := r.URL.Query().Get("job_id")
	if jobIDStr == "" {
		jh.logger.Error("failed to get result", slog.String("error", "job_id is required"))
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	jobID, err := jobs.JobIDFromString(jobIDStr)
	if err != nil {
		jh.logger.Error("failed to parse job id", slog.String("error", err.Error()))
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	resCh, ok := jh.jobStore.Get(jobID)

	if !ok {
		jh.logger.Error("job not found", slog.String("job_id", jobID.String()))
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	select {
	case res := <-resCh:
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(res)); err != nil {
			jh.logger.Error("failed to write result", slog.String("error", err.Error()))
		}
		jh.jobStore.Delete(jobID)

	case <-ctx.Done():
		w.WriteHeader(http.StatusRequestTimeout)
	}
}
