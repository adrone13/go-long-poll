package jobs

import (
	"strconv"
	"sync"
	"sync/atomic"
)

type JobID int64

func (j JobID) String() string {
	return strconv.FormatInt(int64(j), 10)
}

func JobIDFromString(s string) (JobID, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return JobID(i), nil
}

type Job struct {
	done   chan struct{}
	result string
}

// Done is closed once the job completes, waking every concurrent waiter at once.
func (j *Job) Done() <-chan struct{} { return j.done }

// Result is only valid after Done() has fired.
func (j *Job) Result() string { return j.result }

type Store struct {
	items    map[JobID]*Job
	m        sync.RWMutex
	jobIDSeq atomic.Int64
}

func New() *Store {
	return &Store{
		items: make(map[JobID]*Job),
	}
}

func (s *Store) Add() JobID {
	jobID := JobID(s.jobIDSeq.Add(1))

	s.m.Lock()
	defer s.m.Unlock()
	s.items[jobID] = &Job{done: make(chan struct{})}

	return jobID
}

func (s *Store) Get(jobID JobID) (*Job, bool) {
	s.m.RLock()
	defer s.m.RUnlock()
	job, ok := s.items[jobID]

	return job, ok
}

// Complete stores the result and closes Done(), releasing every waiter.
func (s *Store) Complete(jobID JobID, result string) {
	s.m.RLock()
	job, ok := s.items[jobID]
	s.m.RUnlock()
	if !ok {
		return
	}

	job.result = result
	close(job.done)
}

func (s *Store) Delete(jobID JobID) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.items, jobID)
}
