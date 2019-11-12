package ytdlwrapper

import (
	"regexp"
	"strconv"

	"github.com/dustin/go-humanize"
)

var progressExpression = regexp.MustCompile(`\[download\]\s+(\d+\.\d)%\s+of\s+~?(\d+\.\d+)(B|KiB|MiB|GiB|TiB|PiB|EiB|ZiB|YiB)\s+at\s+(\d+\.\d+)(B|KiB|MiB|GiB|TiB|PiB|EiB|ZiB|YiB)\/s\s+ETA\s+(\d\d:\d\d:\d\d|\d\d:\d\d)`)

type DownloadProgress struct {
	Downloaded uint64
	TotalSize  uint64
	Speed      uint64
	Percent    string
	// ETA        time.Duration
}

func sizeUnitToBytes(rawSize []byte, rawUnit []byte) uint64 {
	size, err := humanize.ParseBytes(string(rawSize) + " " + string(rawUnit))

	if err != nil {
		return 0
	}

	return size
}

func parseProgressLine(line []byte) *DownloadProgress {
	// [download]   0.0% of 1.29GiB at  2.61MiB/s ETA 08:28

	matches := progressExpression.FindSubmatch(line)
	if matches == nil {
		return nil
	}

	size := sizeUnitToBytes(matches[2], matches[3])
	speed := sizeUnitToBytes(matches[4], matches[5])

	percent, _ := strconv.ParseFloat(string(matches[1]), 32)
	downloaded := uint64(float64(size)*percent) / 100

	// eta := matches[6]

	return &DownloadProgress{
		Downloaded: downloaded,
		TotalSize:  size,
		Speed:      speed,
		Percent:    string(matches[1]),
	}
}
