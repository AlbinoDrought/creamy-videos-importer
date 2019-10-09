package main

import (
	"errors"
	"sync"
	"time"

	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
)

type jobInformation struct {
	lock sync.RWMutex

	ID     creamqueue.JobID
	Status string

	CreatedAt time.Time
	StartedAt time.Time
	StoppedAt time.Time

	Data     creamqueue.JobData
	Failures []creamqueue.JobFailure
	Result   creamqueue.JobResult
}

type jobRepository struct {
	lock sync.RWMutex

	jobs map[creamqueue.JobID]*jobInformation
}

func makeJobRepository() *jobRepository {
	return &jobRepository{
		jobs: make(map[creamqueue.JobID]*jobInformation),
	}
}

func (repo *jobRepository) Store(id creamqueue.JobID, updater func(job *jobInformation)) error {
	job := &jobInformation{
		ID:       id,
		Failures: []creamqueue.JobFailure{},
	}
	updater(job)

	repo.lock.Lock()
	defer repo.lock.Unlock()
	if _, ok := repo.jobs[id]; ok {
		return errors.New("Job ID already exists: " + string(id))
	}

	repo.jobs[id] = job

	return nil
}

func (repo *jobRepository) Update(id creamqueue.JobID, updater func(job *jobInformation)) error {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	job, ok := repo.jobs[id]
	if !ok {
		return errors.New("Job ID not found: " + string(id))
	}

	job.lock.Lock()
	defer job.lock.Unlock()

	updater(job)

	return nil
}

func (repo *jobRepository) Remove(id creamqueue.JobID) {
	repo.lock.Lock()
	defer repo.lock.Unlock()
	delete(repo.jobs, id)
}

func (repo *jobRepository) Stats() map[string]int {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	stats := map[string]int{}

	for _, job := range repo.jobs {
		job.lock.RLock()
		status := job.Status
		job.lock.RUnlock()

		_, ok := stats[status]
		if !ok { // is this needed? not sure if it just defaults to 0
			stats[status] = 0
		}
		stats[status]++
	}

	return stats
}
