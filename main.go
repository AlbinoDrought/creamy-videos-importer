package main

import (
	"context"
	"log"
	"os"
	"os/signal"

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

	ctx, cancel := context.WithCancel(context.Background())

	workersFinished := bootQueue(ctx)

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
