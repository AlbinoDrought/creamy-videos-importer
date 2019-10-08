package creamqueue

import "context"

// JobID is a unique identifier for a job
type JobID string

// JobData contains the arguments to begin processing our job
type JobData struct {
	URL string
}

// JobResult is the output data of successfully processing a job
type JobResult struct {
	Title     string
	CreamyURL string
}

// JobFailure is the reason why we couldn't process a job
type JobFailure struct {
	Error error
}

// A QueuedJob is something we have to do... eventually
type QueuedJob interface {
	ID() JobID
	Data() *JobData

	Finished(result *JobResult)
	Failed(failure *JobFailure)
}

// An OnQueuedHandler is called when a job is pushed to the queue
type OnQueuedHandler func(id JobID, data JobData)

// An OnStartedHandler is called when a job is started for the first time
type OnStartedHandler func(id JobID, data JobData)

// An OnFinishedHandler is called when a job finishes successfully
type OnFinishedHandler func(id JobID, data JobData, result JobResult)

// An OnFailedHandler is called when a job is unable to finish successfully
type OnFailedHandler func(id JobID, data JobData, failures []JobFailure)

// A Queue handles your jobs
type Queue interface {
	OnQueued(handler OnQueuedHandler)
	OnStarted(handler OnStartedHandler)
	OnFinished(handler OnFinishedHandler)
	OnFailed(handler OnFailedHandler)

	Push(id JobID, data JobData)
	Pull(ctx context.Context) QueuedJob
}
