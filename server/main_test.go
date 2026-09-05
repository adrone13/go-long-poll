package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestConcurrentSubmitAndResult drives concurrent submit+result requests to
// exercise the unsynchronized `jobs` map. Run with -race to catch it:
//
//	go test -race ./server/...
func TestConcurrentSubmitAndResult(t *testing.T) {
	jobSteps = 1
	jobStepDelay = time.Millisecond

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/result", resultHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			submitRes, err := client.Get(srv.URL + "/submit")
			if err != nil {
				t.Errorf("submit: %v", err)
				return
			}
			defer submitRes.Body.Close()

			body, err := io.ReadAll(submitRes.Body)
			if err != nil {
				t.Errorf("read job id: %v", err)
				return
			}
			jobID := string(body)

			resultRes, err := client.Get(fmt.Sprintf("%s/result?job_id=%s", srv.URL, jobID))
			if err != nil {
				t.Errorf("result: %v", err)
				return
			}
			defer resultRes.Body.Close()

			if resultRes.StatusCode != http.StatusOK {
				t.Errorf("result status = %d, want %d", resultRes.StatusCode, http.StatusOK)
			}
		}()
	}
	wg.Wait()
}
