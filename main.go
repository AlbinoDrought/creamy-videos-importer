package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/AlbinoDrought/creamy-videos-importer/autoid"
	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
	"github.com/AlbinoDrought/creamy-videos-importer/creamyvideos"
	"github.com/AlbinoDrought/creamy-videos-importer/ytdlwrapper"
)

var queue creamqueue.Queue
var idGenerator autoid.AutoID
var creamyVideosHost string
var parallelWorkers int

func main() {
	queue = creamqueue.MakeBarebonesQueue()
	idGenerator = autoid.Make()
	creamyVideosHost = "http://localhost:3000/"
	parallelWorkers = 3

	queue.OnFinished(func(id creamqueue.JobID, data creamqueue.JobData, result creamqueue.JobResult) {
		log.Println("finished", id, data.URL, result.Title, result.CreamyURL)
	})

	queue.OnFailed(func(id creamqueue.JobID, data creamqueue.JobData, failures []creamqueue.JobFailure) {
		log.Println("failed", id, data.URL, failures)
	})

	queue.OnStarted(func(id creamqueue.JobID, data creamqueue.JobData) {
		log.Println("started", id, data.URL)
	})

	queue.OnQueued(func(id creamqueue.JobID, data creamqueue.JobData) {
		log.Println("queued", id, data.URL)
	})

	ctx, cancel := context.WithCancel(context.Background())

	workerWaitGroup := sync.WaitGroup{}
	for i := 0; i < parallelWorkers; i++ {
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

func workQueue(ctx context.Context) {
	var job creamqueue.QueuedJob
	for {
		job = queue.Pull(ctx)
		if job == nil {
			return
		}
		processJob(ctx, job)
	}
}

func processJob(ctx context.Context, job creamqueue.QueuedJob) {
	jobData := job.Data()
	url := jobData.URL
	wrapper := ytdlwrapper.Make()

	info, err := wrapper.Info(ctx, url)
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	if info.IsPlaylist {
		for _, entry := range info.Playlist.Entries {
			queue.Push(idGenerator.Next(), creamqueue.JobData{
				URL:                     entry.URL,
				ParentPlaylistID:        info.Playlist.ID,
				ParentPlaylistExtractor: info.Playlist.Extractor,
			})
		}

		job.Finished(&creamqueue.JobResult{
			Title: "Playlist " + info.Playlist.ID,
		})
		return
	}

	outputFilenameBytes, err := wrapper.Download(ctx, info.Entry.URL, "--get-filename", "-f", "best[ext=mp4]/best[ext=webm]/best", "-o", string(job.ID())+".%(ext)s")
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	outputFilename := strings.TrimSpace(string(outputFilenameBytes))
	defer os.Remove(outputFilename)
	defer os.Remove(outputFilename + ".part")

	_, err = wrapper.Download(ctx, info.Entry.URL, "-f", "best[ext=mp4]/best[ext=webm]/best", "-o", outputFilename)
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	title := info.Entry.Title
	if title == "" {
		title = "Import of " + url
	}

	description := "Original URL: " + url
	if info.Entry.Description != "" {
		description += "\n\n" + info.Entry.Description
	}

	tags := []string{"importer:cvi"}

	if info.Entry.Extractor != "" {
		tags = append(tags, "extractor:"+info.Entry.Extractor)

		if info.Entry.ChannelID != "" {
			tags = append(tags, fmt.Sprintf("%v-channel:%v", info.Entry.Extractor, info.Entry.ChannelID))
		}

		if info.Entry.UploaderID != "" {
			tags = append(tags, fmt.Sprintf("%v-uploader:%v", info.Entry.Extractor, info.Entry.UploaderID))
		}

		if info.Entry.ID != "" {
			tags = append(tags, fmt.Sprintf("%v-id:%v", info.Entry.Extractor, info.Entry.ID))
		}
	}

	if jobData.ParentPlaylistID != "" {
		if jobData.ParentPlaylistExtractor != "" {
			extractor := strings.Replace(jobData.ParentPlaylistExtractor, ":playlist", "", -1)
			tags = append(tags, fmt.Sprintf("%v-playlist:%v", extractor, jobData.ParentPlaylistID))
		} else {
			tags = append(tags, "imported-playlist:"+jobData.ParentPlaylistID)
		}
	}

	result, err := creamyvideos.Upload(
		creamyVideosHost,
		outputFilename,
		title,
		description,
		tags,
	)

	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	job.Finished(&creamqueue.JobResult{
		Title:     info.Entry.Title,
		CreamyURL: result.URL,
	})
}
