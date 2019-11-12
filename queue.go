package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
)

func bootQueue(ctx context.Context) chan bool {
	queue.OnFinished(func(id creamqueue.JobID, data creamqueue.JobData, result creamqueue.JobResult) {
		log.Println("finished", id, data.URL, result.Title, result.CreamyURL)
		jobRepo.Update(id, func(job *jobInformation) {
			job.StoppedAt = time.Now()
			job.Status = "finished"
			job.Data = data
			job.Result = result
		})
	})

	queue.OnFailed(func(id creamqueue.JobID, data creamqueue.JobData, failures []creamqueue.JobFailure) {
		log.Println("failed", id, data.URL, failures)
		jobRepo.Update(id, func(job *jobInformation) {
			job.StoppedAt = time.Now()
			job.Status = "failed"
			job.Data = data
			job.Failures = failures
		})
	})

	queue.OnStarted(func(id creamqueue.JobID, data creamqueue.JobData) {
		log.Println("started", id, data.URL)
		jobRepo.Update(id, func(job *jobInformation) {
			job.StartedAt = time.Now()
			job.Status = "started"
			job.Data = data
		})
	})

	queue.OnProgress(func(id creamqueue.JobID, data creamqueue.JobData, progress creamqueue.JobProgress) {
		log.Println("progress", id, progress)
		jobRepo.Update(id, func(job *jobInformation) {
			job.Progress = progress
		})
	})

	queue.OnQueued(func(id creamqueue.JobID, data creamqueue.JobData) {
		log.Println("queued", id, data.URL)
		jobRepo.Store(id, func(job *jobInformation) {
			job.CreatedAt = time.Now()
			job.Status = "waiting"
			job.Data = data
		})
	})

	workerWaitGroup := sync.WaitGroup{}
	for i := 0; i < config.parallelWorkers; i++ {
		workerWaitGroup.Add(1)
		go func() {
			workQueue(ctx)
			workerWaitGroup.Done()
		}()
	}

	workersFinished := make(chan bool, 1)
	go func() {
		workerWaitGroup.Wait()
		workersFinished <- true
	}()

	return workersFinished
}
