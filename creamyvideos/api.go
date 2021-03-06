package creamyvideos

import (
	"encoding/json"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/imroc/req"
)

const pointUploadVideo = "/api/upload"
const pointWatchVideo = "/watch/"

// UploadResult is the result of uploading a video
type UploadResult struct {
	ID  string
	URL string
}

func point(host, path string) (string, error) {
	parsedHost, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	parsedHost.Path = path
	return parsedHost.String(), nil
}

// UploadWithProgress uploads a local file to a creamy-videos server and provides progress updates
func UploadWithProgress(host, localPath, title, description string, tags []string, callback func(current, total int64)) (*UploadResult, error) {
	r := req.New()

	url, err := point(host, pointUploadVideo)
	if err != nil {
		return nil, err
	}

	fileStream, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer fileStream.Close()

	resp, err := r.Post(
		url,
		req.Param{
			"title":       title,
			"description": description,
			"tags":        strings.Join(tags, ","),
		},
		req.FileUpload{
			File:      fileStream,
			FieldName: "file",
			FileName:  path.Base(localPath),
		},
		req.UploadProgress(callback),
	)

	if err != nil {
		return nil, err
	}

	responseBody := struct {
		ID uint64 `json:"id"`
	}{}

	err = json.NewDecoder(resp.Response().Body).Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	stringID := strconv.FormatUint(responseBody.ID, 10)

	watchURL, err := point(host, pointWatchVideo+stringID)

	uploadResult := &UploadResult{
		ID:  stringID,
		URL: watchURL,
	}

	return uploadResult, err
}

// Upload a local file to a creamy-videos server
func Upload(host, localPath, title, description string, tags []string) (*UploadResult, error) {
	return UploadWithProgress(host, localPath, title, description, tags, func(current, total int64) {})
}
