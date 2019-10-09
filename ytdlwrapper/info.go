package ytdlwrapper

// An Entry is a single video
type Entry struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	UploadDate  string `json:"upload_date"`
	UploaderID  string `json:"uploader_id"`
	ChannelID   string `json:"channel_id"`
	Description string `json:"description"`
	Extractor   string `json:"extractor"`
	URL         string `json:"webpage_url"`
}

// A Playlist is a list of entries
type Playlist struct {
	ID        string  `json:"id"`
	Extractor string  `json:"extractor"`
	Entries   []Entry `json:"entries"`
}

// InfoOutput represents the JSON returned by `youtube-dl -J`.
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
