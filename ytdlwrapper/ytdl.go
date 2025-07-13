package ytdlwrapper

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
)

// A Wrapper for the youtube-dl or yt-dlp binary
type Wrapper struct {
	BinPath string
}

// Info returns information about the URL, like if it is
// a playlist or a single video.
func (wrapper *Wrapper) Info(ctx context.Context, url string) (*InfoOutput, error) {
	output, err := exec.CommandContext(ctx, wrapper.BinPath, "-J", "--flat-playlist", "--no-playlist", url).Output()
	if err != nil {
		return nil, err
	}

	unknown := unknownInfo{}
	err = json.Unmarshal(output, &unknown)
	if err != nil {
		return nil, err
	}

	infoOutput := InfoOutput{}
	if unknown.Type == "playlist" {
		err = json.Unmarshal(output, &infoOutput.Playlist)
		infoOutput.IsPlaylist = true
	} else {
		err = json.Unmarshal(output, &infoOutput.Entry)
		infoOutput.IsPlaylist = false
	}

	return &infoOutput, err
}

// Update youtube-dl or yt-dlp
func (wrapper *Wrapper) Update(ctx context.Context) error {
	_, err := exec.CommandContext(ctx, wrapper.BinPath, "-U").Output()
	return err
}

// Download the given URL using youtube-dl or yt-dlp
func (wrapper *Wrapper) Download(ctx context.Context, url string, args ...string) ([]byte, error) {
	args = append(args, url)
	return exec.CommandContext(ctx, wrapper.BinPath, args...).Output()
}

// DownloadWithProgress downloads the given URL using youtube-dl or yt-dlp and provides progress updates
func (wrapper *Wrapper) DownloadWithProgress(ctx context.Context, callback func(*DownloadProgress), url string, args ...string) error {
	args = append(args, "--newline", url)
	cmd := exec.CommandContext(ctx, wrapper.BinPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			rawProgress := scanner.Bytes()
			parsedProgress := parseProgressLine(rawProgress)
			if parsedProgress != nil {
				callback(parsedProgress)
			}
		}
	}()

	return cmd.Wait()
}

// Make a default instance of the youtube-dl wrapper
func Make() *Wrapper {
	binPath := os.Getenv("CREAMY_YTDL_BIN_PATH")
	if binPath == "" {
		binPath = "youtube-dl"
	}
	return &Wrapper{
		BinPath: binPath,
	}
}
