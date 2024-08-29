// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	progressPrefix     = "dl:"
	progressSplitChars = "###"
)

var progressTemplateFields = []string{
	"%(progress.status)s",
	"%(progress.total_bytes,progress.total_bytes_estimate)s",
	"%(progress.downloaded_bytes)s",
	"%(progress.fragment_index)s",
	"%(progress.fragment_count)s",
	"%(info.id)s",
	"%(info.playlist_id)s",
	"%(info.playlist_index)s",
	"%(info.playlist_count)s",
	"%(info.url,info.webpage_url)s",
	"%(progress.filename,progress.tmpfilename,info.filename)s",
}

func strFloatToInt(s string) int {
	f, _ := strconv.ParseFloat(s, 64)
	return int(f)
}

type progressHandler struct {
	fn ProgressCallbackFunc

	mu       sync.Mutex
	started  map[string]time.Time // Used to track multiple independent downloads.
	finished map[string]time.Time // Used to track multiple independent downloads.
}

func newProgressHandler(fn ProgressCallbackFunc) *progressHandler {
	h := &progressHandler{
		fn:       fn,
		started:  make(map[string]time.Time),
		finished: make(map[string]time.Time),
	}
	return h
}

func (h *progressHandler) parse(input string) {
	raw := strings.SplitN(input, progressSplitChars, len(progressTemplateFields))

	if len(raw) != len(progressTemplateFields) {
		return
	}

	// Clean up the raw data.
	for i := range raw {
		raw[i] = strings.TrimSpace(raw[i])
		if raw[i] == "NA" {
			raw[i] = ""
		}
	}

	update := ProgressUpdate{
		Status:          ProgressStatus(raw[0]),
		TotalBytes:      strFloatToInt(raw[1]), // Total bytes, if available.
		DownloadedBytes: strFloatToInt(raw[2]),
		FragmentIndex:   strFloatToInt(raw[3]),
		FragmentCount:   strFloatToInt(raw[4]),
		ID:              raw[5],
		PlaylistID:      raw[6],
		PlaylistIndex:   strFloatToInt(raw[7]),
		PlaylistCount:   strFloatToInt(raw[8]),
		URL:             raw[9],
		Filename:        raw[10],
	}

	uuid := update.uuid()

	var ok bool

	h.mu.Lock()
	update.Started, ok = h.started[uuid]
	if !ok {
		update.Started = time.Now()
		h.started[uuid] = update.Started
	}

	update.Finished, ok = h.finished[uuid]
	if !ok && update.Status.IsCompletedType() {
		update.Finished = time.Now()
		h.finished[uuid] = update.Finished
	}
	h.mu.Unlock()

	h.fn(update)
}

// ProgressStatus is the status of the download progress.
type ProgressStatus string

func (s ProgressStatus) IsCompletedType() bool {
	return s == ProgressStatusError || s == ProgressStatusFinished
}

const (
	ProgressStatusStarting       ProgressStatus = "starting"
	ProgressStatusDownloading    ProgressStatus = "downloading"
	ProgressStatusPostProcessing ProgressStatus = "post_processing"
	ProgressStatusError          ProgressStatus = "error"
	ProgressStatusFinished       ProgressStatus = "finished"
)

// ProgressCallbackFunc is a callback function that is called when (if) we receive
// progress updates from yt-dlp.
type ProgressCallbackFunc func(update ProgressUpdate)

// ProgressUpdate is a point-in-time snapshot of the download progress.
type ProgressUpdate struct {
	// Status is the current status of the download.
	Status ProgressStatus
	// TotalBytes is the total number of bytes in the download. If yt-dlp is unable
	// to determine the total bytes, this will be 0.
	TotalBytes int
	// DownloadedBytes is the number of bytes that have been downloaded so far.
	DownloadedBytes int
	// FragmentIndex is the index of the current fragment being downloaded.
	FragmentIndex int
	// FragmentCount is the total number of fragments in the download.
	FragmentCount int

	// ID is the ID of the video being downloaded if available.
	ID string
	// PlaylistID is the ID of the playlist the video is in, if available.
	PlaylistID string
	// PlaylistIndex is the index of the video in the playlist, if available.
	PlaylistIndex int
	// PlaylistCount is the total number of videos in the playlist, if available.
	PlaylistCount int
	// URL is the URL of the video being downloaded.
	URL string
	// Filename is the filename of the video being downloaded, if available. Note that
	// this is not necessarily the same as the destination file, as post-processing
	// may merge multiple files into one.
	Filename string

	// Started is the time the download started.
	Started time.Time
	// Finished is the time the download finished. If the download is still in progress,
	// this will be zero. You can validate with IsZero().
	Finished time.Time
}

func (p *ProgressUpdate) uuid() string {
	return strings.Join([]string{
		p.ID,
		p.PlaylistID,
		strconv.Itoa(p.PlaylistIndex),
		p.Filename,
	}, ":")
}

// Duration returns the duration of the download. If the download is still in progress,
// it will return the time since the download started.
func (p *ProgressUpdate) Duration() time.Duration {
	if p.Finished.IsZero() {
		return time.Since(p.Started)
	}
	return p.Finished.Sub(p.Started)
}

// ETA returns the estimated time until the download is complete. If the download is
// complete, or hasn't started yet, it will return 0.
func (p *ProgressUpdate) ETA() time.Duration {
	perc := p.Percent()
	if perc == 0 || perc == 100 {
		return 0
	}
	return time.Duration(float64(p.Duration().Nanoseconds()) / perc * (100 - perc))
}

// Percent returns the percentage of the download that has been completed. If yt-dlp
// is unable to determine the total bytes, it will return 0.
func (p *ProgressUpdate) Percent() float64 {
	if p.Status.IsCompletedType() {
		return 100
	}
	if p.TotalBytes == 0 {
		return 0
	}
	return float64(p.DownloadedBytes) / float64(p.TotalBytes) * 100
}

// PercentString is like Percent, but returns a string representation of the percentage.
func (p *ProgressUpdate) PercentString() string {
	return fmt.Sprintf("%.2f%%", p.Percent())
}

// ProgressFunc can be used to register a callback function that will be called when
// yt-dlp sends progress updates. The callback function will be called with any information
// that yt-dlp is able to provide, including sending separate updates for each file, playlist,
// etc that may be downloaded.
//   - See [Command.UnsetProgressFunc], for unsetting the progress function.
func (c *Command) ProgressFunc(frequency time.Duration, fn ProgressCallbackFunc) *Command {
	if frequency < 100*time.Millisecond {
		frequency = 100 * time.Millisecond
	}

	c.Progress().
		ProgressDelta(frequency.Seconds()).
		ProgressTemplate(progressPrefix + strings.Join(progressTemplateFields, progressSplitChars)).
		Newline()

	c.mu.Lock()
	c.progress = newProgressHandler(fn)
	c.mu.Unlock()

	return c
}

// UnsetProgressFunc can be used to unset the progress function that was previously set
// with [Command.ProgressFunc].
func (c *Command) UnsetProgressFunc() *Command {
	c.mu.Lock()
	c.progress = nil
	c.mu.Unlock()

	return c
}
