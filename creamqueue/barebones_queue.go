package creamqueue

import (
	"context"
	"sync"
)

type barebonesJob struct {
	id    JobID
	queue *barebonesQueue

	attempts         uint
	maxAttempts      uint
	previouslyPulled bool
	failures         []JobFailure

	data *JobData
}

func (job *barebonesJob) ID() JobID {
	return job.id
}

func (job *barebonesJob) Data() *JobData {
	return job.data
}

func (job *barebonesJob) Finished(result *JobResult) {
	go job.queue.triggerFinished(job, result)
}

func (job *barebonesJob) Failed(failure *JobFailure) {
	go job.queue.triggerFailed(job, failure)
}

type barebonesQueue struct {
	jobs         chan *barebonesJob
	priorityJobs chan *barebonesJob

	handlerLock      sync.Locker
	queuedHandlers   []OnQueuedHandler
	startedHandlers  []OnStartedHandler
	finishedHandlers []OnFinishedHandler
	failedHanders    []OnFailedHandler
}

func (queue *barebonesQueue) OnQueued(handler OnQueuedHandler) {
	queue.handlerLock.Lock()
	queue.queuedHandlers = append(queue.queuedHandlers, handler)
	queue.handlerLock.Unlock()
}

func (queue *barebonesQueue) triggerQueued(job *barebonesJob) {
	for _, handler := range queue.queuedHandlers {
		handler(job.id, *job.data)
	}
}

func (queue *barebonesQueue) OnStarted(handler OnStartedHandler) {
	queue.handlerLock.Lock()
	queue.startedHandlers = append(queue.startedHandlers, handler)
	queue.handlerLock.Unlock()
}

func (queue *barebonesQueue) triggerStarted(job *barebonesJob) {
	for _, handler := range queue.startedHandlers {
		handler(job.id, *job.data)
	}
}

func (queue *barebonesQueue) OnFinished(handler OnFinishedHandler) {
	queue.handlerLock.Lock()
	queue.finishedHandlers = append(queue.finishedHandlers, handler)
	queue.handlerLock.Unlock()
}

func (queue *barebonesQueue) triggerFinished(job *barebonesJob, result *JobResult) {
	for _, handler := range queue.finishedHandlers {
		handler(job.id, *job.data, *result)
	}
}

func (queue *barebonesQueue) OnFailed(handler OnFailedHandler) {
	queue.handlerLock.Lock()
	queue.failedHanders = append(queue.failedHanders, handler)
	queue.handlerLock.Unlock()
}

func (queue *barebonesQueue) triggerFailed(job *barebonesJob, failure *JobFailure) {
	if job.attempts < job.maxAttempts {
		job.attempts++
		job.failures = append(job.failures, *failure)
		go queue.pushToPriorityQueue(job)
		return
	}

	for _, handler := range queue.failedHanders {
		handler(job.id, *job.data, job.failures)
	}
}

func (queue *barebonesQueue) pushToQueue(job *barebonesJob) {
	queue.jobs <- job
}

func (queue *barebonesQueue) pushToPriorityQueue(job *barebonesJob) {
	queue.priorityJobs <- job
}

func (queue *barebonesQueue) Push(id JobID, data JobData) {
	job := &barebonesJob{
		id:    id,
		queue: queue,

		attempts:    0,
		maxAttempts: 2,
		failures:    []JobFailure{},

		data: &data,
	}

	queue.triggerQueued(job)
	go queue.pushToQueue(job)
}

func (queue *barebonesQueue) Pull(ctx context.Context) QueuedJob {
	var job *barebonesJob

	select {
	case <-ctx.Done():
		return nil
	// prefer the priority queue:
	case job = <-queue.priorityJobs:
		break
	case job = <-queue.jobs:
		break
	}

	if !job.previouslyPulled {
		queue.triggerStarted(job)
		job.previouslyPulled = true
	}

	return job
}

// MakeBarebonesQueue returns a perfectly valid and working Queue instance :^)
func MakeBarebonesQueue() Queue {
	return &barebonesQueue{
		jobs:             make(chan *barebonesJob),
		priorityJobs:     make(chan *barebonesJob),
		handlerLock:      &sync.Mutex{},
		queuedHandlers:   []OnQueuedHandler{},
		startedHandlers:  []OnStartedHandler{},
		finishedHandlers: []OnFinishedHandler{},
		failedHanders:    []OnFailedHandler{},
	}
}
