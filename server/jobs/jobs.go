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

type Store struct {
	items    map[JobID]chan string
	m        sync.RWMutex
	jobIDSeq atomic.Int64
}

func New() *Store {
	return &Store{
		items: make(map[JobID]chan string),
	}
}

func (j *Store) Add() JobID {
	jobID := JobID(j.jobIDSeq.Add(1))

	j.m.Lock()
	defer j.m.Unlock()
	j.items[jobID] = make(chan string, 1)

	return jobID
}

func (j *Store) Get(jobID JobID) (chan string, bool) {
	j.m.RLock()
	defer j.m.RUnlock()
	ch, ok := j.items[jobID]

	return ch, ok
}

func (j *Store) Delete(jobID JobID) {
	j.m.Lock()
	defer j.m.Unlock()
	delete(j.items, jobID)
}
