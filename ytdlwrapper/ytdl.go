package ytdlwrapper

import (
	"context"
	"encoding/json"
	"os/exec"
)

// A Wrapper for the youtube-dl binary
type Wrapper struct {
	BinPath string
}

// Info returns information about the URL, like if it is
// a playlist or a single video.
func (wrapper *Wrapper) Info(ctx context.Context, url string) (*InfoOutput, error) {
	output, err := exec.CommandContext(ctx, wrapper.BinPath, "-J", url).Output()
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

// Update youtube-dl
func (wrapper *Wrapper) Update(ctx context.Context) error {
	_, err := exec.CommandContext(ctx, wrapper.BinPath, "-U").Output()
	return err
}

// Download the given URL using youtube-dl
func (wrapper *Wrapper) Download(ctx context.Context, url string, args ...string) ([]byte, error) {
	args = append(args, url)
	return exec.CommandContext(ctx, wrapper.BinPath, args...).Output()
}

// Make a default instance of the youtube-dl wrapper
func Make() *Wrapper {
	return &Wrapper{
		BinPath: "youtube-dl",
	}
}
