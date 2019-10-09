package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

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

	gracefulWaitGroup := sync.WaitGroup{}
	gracefulShutdownComplete := make(chan bool, 1)

	workersFinished := bootQueue(ctx)
	gracefulWaitGroup.Add(1)
	go func() {
		<-workersFinished
		gracefulWaitGroup.Done()
	}()

	serverFinished := bootServer(ctx)
	gracefulWaitGroup.Add(1)
	go func() {
		err := <-serverFinished
		if err != nil {
			log.Println("server exited with error", err)
		}
		gracefulWaitGroup.Done()
	}()

	go func() {
		gracefulWaitGroup.Wait()
		gracefulShutdownComplete <- true
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	firstInterrupt := true

	for {
		select {
		case <-gracefulShutdownComplete:
			log.Println("Graceful shutdown finished, bye!")
			return
		case <-c:
			if firstInterrupt {
				firstInterrupt = false
				cancel()
				log.Println("Interrupt received, initiated graceful shutdown")
			} else {
				log.Println("Performing unclean shutdown")
				return
			}
		}
	}
}
