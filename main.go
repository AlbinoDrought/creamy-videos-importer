package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/AlbinoDrought/creamy-videos-importer/autoid"
	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
)

var queue creamqueue.Queue
var idGenerator autoid.AutoID
var jobRepo *jobRepository

var config = struct {
	creamyVideosHost string
	parallelWorkers  int
}{}

func main() {
	queue = creamqueue.MakeBarebonesQueue()
	idGenerator = autoid.Make()
	jobRepo = makeJobRepository()

	config.creamyVideosHost = "http://localhost:3000/"
	config.parallelWorkers = 3

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

	queue.OnQueued(func(id creamqueue.JobID, data creamqueue.JobData) {
		log.Println("queued", id, data.URL)
		jobRepo.Store(id, func(job *jobInformation) {
			job.CreatedAt = time.Now()
			job.Status = "waiting"
			job.Data = data
		})
	})

	ctx, cancel := context.WithCancel(context.Background())

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

	queue.Push(idGenerator.Next(), creamqueue.JobData{
		URL: "https://www.youtube.com/playlist?list=PLkxPfMNWejkdjjBA4PruQz5oyPIxCfeF7",
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	firstInterrupt := true

	for {
		select {
		case <-workersFinished:
			log.Println("All workers finished, bye!")
			return
		case <-c:
			if firstInterrupt {
				log.Println("Interrupt received, waiting for workers to finish cleanly")
				firstInterrupt = false
				cancel()
			} else {
				log.Println("Performing unclean shutdown")
				return
			}
		}
	}
}
