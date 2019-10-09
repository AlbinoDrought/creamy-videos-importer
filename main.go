package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AlbinoDrought/creamy-videos-importer/autoid"
	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
	"github.com/AlbinoDrought/creamy-videos-importer/ytdlwrapper"
)

var queue creamqueue.Queue
var idGenerator autoid.AutoID

func main() {
	queue = creamqueue.MakeBarebonesQueue()
	idGenerator = autoid.Make()

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

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		go workQueue(ctx)
	}

	queue.Push(idGenerator.Next(), creamqueue.JobData{
		URL: "https://www.youtube.com/playlist?list=PLkxPfMNWejkdjjBA4PruQz5oyPIxCfeF7",
	})

	fmt.Scanln()
}

func workQueue(ctx context.Context) {
	var job creamqueue.QueuedJob
	for {
		job = queue.Pull(ctx)
		if job == nil {
			return
		}
		processJob(job)
	}
}

func processJob(job creamqueue.QueuedJob) {
	url := job.Data().URL
	wrapper := ytdlwrapper.Make()

	info, err := wrapper.Info(url)
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
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

	_, err = wrapper.Download(info.Entry.URL, "-f", "best[ext=mp4]/best[ext=webm]/best", "-o", string(job.ID())+".%(ext)s")
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	outputFilenameBytes, err := wrapper.Download(info.Entry.URL, "--get-filename", "-f", "best[ext=mp4]/best[ext=webm]/best", "-o", string(job.ID())+".%(ext)s")
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	outputFilename := string(outputFilenameBytes)
	defer os.Remove(outputFilename)

	job.Finished(&creamqueue.JobResult{
		Title:     info.Entry.Title,
		CreamyURL: "none because PoC",
	})
}
