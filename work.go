package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
	"github.com/AlbinoDrought/creamy-videos-importer/creamyvideos"
	"github.com/AlbinoDrought/creamy-videos-importer/ytdlwrapper"
	"github.com/dustin/go-humanize"
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

	job.Progress(creamqueue.JobProgress("Fetching info"))
	info, err := wrapper.Info(ctx, url)
	if err != nil {
		job.Progress(creamqueue.JobProgress("Failed fetching info"))
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	if info.IsPlaylist {
		if jobData.ParentPlaylistID != "" {
			// https://github.com/AlbinoDrought/creamy-videos-importer/issues/11
			// Some playlists try to re-import themselves, leading to an endless loop.
			// To fix this, abort import if a playlist tries to import a playlist.
			job.Progress(creamqueue.JobProgress("Job triggered by playlist import is another playlist! Aborting"))
			job.Failed(&creamqueue.JobFailure{
				Error: fmt.Errorf(
					"Job triggered by parent playlist %s tried to import another playlist %s, aborting",
					jobData.ParentPlaylistID,
					info.Playlist.ID,
				),
			})
			return
		}

		for _, entry := range info.Playlist.Entries {
			queue.Push(idGenerator.Next(), creamqueue.JobData{
				URL:                     entry.BestURL(),
				Tags:                    tags,
				ParentPlaylistID:        info.Playlist.ID,
				ParentPlaylistExtractor: info.Playlist.Extractor,
			})
		}

		job.Progress(creamqueue.JobProgress("Queued child videos!"))
		job.Finished(&creamqueue.JobResult{
			Title: "Playlist " + info.Playlist.ID,
		})
		return
	}

	entryURL := info.Entry.BestURL()

	job.Progress(creamqueue.JobProgress("Fetching output filename"))
	outputFilenameBytes, err := wrapper.Download(ctx, entryURL, "--no-playlist", "--get-filename", "-f", "best[ext=mp4]/best[ext=webm]/best/mp4/webm", "-o", string(job.ID())+".%(ext)s")
	if err != nil {
		job.Progress(creamqueue.JobProgress("Failed fetching output filename"))
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

	job.Progress(creamqueue.JobProgress("Starting download"))
	downloadProgressCallback := func(progress *ytdlwrapper.DownloadProgress) {
		job.Progress(creamqueue.JobProgress(fmt.Sprintf(
			"Download %v%% complete (downloaded %v / %v @ %v/s)",
			progress.Percent,
			humanize.Bytes(progress.Downloaded),
			humanize.Bytes(progress.TotalSize),
			humanize.Bytes(progress.Speed),
		)))
	}

	err = wrapper.DownloadWithProgress(ctx, downloadProgressCallback, entryURL, "--no-playlist", "-f", "best[ext=mp4]/best[ext=webm]/best/mp4/webm", "-o", outputFilename)
	if err != nil {
		job.Progress(creamqueue.JobProgress("Failed downloading"))
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

	job.Progress(creamqueue.JobProgress("Uploading"))
	uploadProgressCallback := func(current, total int64) {
		job.Progress(creamqueue.JobProgress(fmt.Sprintf(
			"Upload %.1f%% complete (uploaded %v / %v)",
			(float32(current) / float32(total) * 100),
			humanize.Bytes(uint64(current)),
			humanize.Bytes(uint64(total)),
		)))
	}
	result, err := creamyvideos.UploadWithProgress(
		config.creamyVideosHost,
		outputFilename,
		title,
		description,
		tags,
		uploadProgressCallback,
	)

	if err != nil {
		job.Progress(creamqueue.JobProgress("Failed uploading"))
		job.Failed(&creamqueue.JobFailure{
			Error: err,
		})
		return
	}

	job.Progress(creamqueue.JobProgress("Uploaded!"))
	job.Finished(&creamqueue.JobResult{
		Title:     info.Entry.Title,
		CreamyURL: result.URL,
	})
}
