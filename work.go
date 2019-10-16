package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
	"github.com/AlbinoDrought/creamy-videos-importer/creamyvideos"
	"github.com/AlbinoDrought/creamy-videos-importer/ytdlwrapper"
)

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
	tags := jobData.Tags
	wrapper := ytdlwrapper.Make()
	checkForTagMatch := false

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
				URL:                     entry.BestURL(),
				Tags:                    tags,
				ParentPlaylistID:        info.Playlist.ID,
				ParentPlaylistExtractor: info.Playlist.Extractor,
			})
		}

		job.Finished(&creamqueue.JobResult{
			Title: "Playlist " + info.Playlist.ID,
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

	tags = append(tags, "importer:cvi")

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
			// an extractor ID is assumed to be unique.
			// if we have information on this, trigger a check for tag matches
			// so we don't upload a ton of dupes:
			checkForTagMatch = true
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

	// check if an existing video has all of our tags.
	// if so, use that instead of reimporting ours again.
	if checkForTagMatch {
		existingResult, err := creamyvideos.FirstForTags(config.creamyVideosHost, tags)
		if err != nil {
			job.Failed(&creamqueue.JobFailure{
				Error: err,
			})
			return
		}

		if existingResult != nil {
			job.Finished(&creamqueue.JobResult{
				Title:     info.Entry.Title,
				CreamyURL: existingResult.URL,
			})
			return
		}
	}

	entryURL := info.Entry.BestURL()

	outputFilenameBytes, err := wrapper.Download(ctx, entryURL, "--no-playlist", "--get-filename", "-f", "best[ext=mp4]/best[ext=webm]/best", "-o", string(job.ID())+".%(ext)s")
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	outputFilename := strings.TrimSpace(string(outputFilenameBytes))

	// cleanup any files now, and also queue their cleanup for later:
	os.Remove(outputFilename)
	defer os.Remove(outputFilename)
	os.Remove(outputFilename + ".part")
	defer os.Remove(outputFilename + ".part")

	_, err = wrapper.Download(ctx, entryURL, "--no-playlist", "-f", "best[ext=mp4]/best[ext=webm]/best", "-o", outputFilename)
	if err != nil {
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	result, err := creamyvideos.Upload(
		config.creamyVideosHost,
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
