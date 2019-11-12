package ytdlwrapper

import (
	"math"
	"regexp"
	"strconv"
)

var progressExpression = regexp.MustCompile(`\[download\]\s+(\d+\.\d)%\s+of\s+~?(\d+\.\d+)(B|KiB|MiB|GiB|TiB|PiB|EiB|ZiB|YiB)\s+at\s+(\d+\.\d+)(B|KiB|MiB|GiB|TiB|PiB|EiB|ZiB|YiB)\/s\s+ETA\s+(\d\d:\d\d:\d\d|\d\d:\d\d)`)

type DownloadProgress struct {
	Downloaded uint64
	TotalSize  uint64
	Speed      uint64
	Percent    string
	// ETA        time.Duration
}

var unitMap = map[string]float64{}

func init() {
	// https://github.com/ytdl-org/youtube-dl/blob/1a01639bf9514c20d54e7460ba9b493b3283ca9a/youtube_dl/utils.py#L1671
	for i, unit := range []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"} {
		unitMap[unit] = math.Pow(1024, float64(i))
	}
}

func sizeUnitToBytes(rawSize []byte, rawUnit []byte) uint64 {
	parsedSize, _ := strconv.ParseFloat(string(rawSize), 64)
	parsedUnit := string(rawUnit)
	unitSize, ok := unitMap[parsedUnit]

	if !ok {
		unitSize = 0
	}

	return uint64(parsedSize * unitSize)
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
	downloaded := uint64(float64(size) * percent)

	// eta := matches[6]

	return &DownloadProgress{
		Downloaded: downloaded,
		TotalSize:  size,
		Speed:      speed,
		Percent:    string(matches[1]),
	}
}
