package ytdlwrapper

import "strings"

// An Entry is a single video
type Entry struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	UploadDate  string `json:"upload_date"`
	UploaderID  string `json:"uploader_id"`
	ChannelID   string `json:"channel_id"`
	Description string `json:"description"`
	Extractor   string `json:"extractor"`
	WebpageURL  string `json:"webpage_url"` // sometimes not set

	// these are set for "URL"-type objects, returned from --flat-playlist
	RawURL string `json:"url"`
	IEKey  string `json:"ie_key"`
}

// BestURL returns the most appropriate URL for an entry
func (entry *Entry) BestURL() string {
	if entry.WebpageURL != "" {
		return entry.WebpageURL
	}

	rawURL := entry.RawURL

	// for some reason, RawURL is just the video ID for Youtube videos.
	// it will be something like this:
	// - RawURL: "wmbpOb9neLY"
	// - IEKey:  "Youtube"
	// we want it to be:
	// https://www.youtube.com/watch?v=wmbpOb9neLY
	if !strings.Contains(rawURL, ":") && entry.IEKey == "Youtube" {
		rawURL = "https://www.youtube.com/watch?v=" + rawURL
	}

	return rawURL
}

// A Playlist is a list of entries
type Playlist struct {
	ID        string  `json:"id"`
	Extractor string  `json:"extractor"`
	Entries   []Entry `json:"entries"`
}

// InfoOutput represents the JSON returned by `youtube-dl -J` or `yt-dlp -J`.
// It changes format depending on if the target is a video
// or an entire playlist, which is why the struct looks like this:
type InfoOutput struct {
	Entry      Entry
	Playlist   Playlist
	IsPlaylist bool
}

// GetAllEntries returns an array of all entries, even if
// InfoOutput is just for a single video.
func (infoOutput *InfoOutput) GetAllEntries() []Entry {
	if infoOutput.IsPlaylist {
		return infoOutput.Playlist.Entries
	}

	return []Entry{infoOutput.Entry}
}

type unknownInfo struct {
	Type string `json:"_type"`
}
