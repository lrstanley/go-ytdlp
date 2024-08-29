// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var progressPrefix = []byte("progress:")

const progressFormat = "%()j"

type progressData struct {
	Info     *ExtractedInfo `json:"info"`
	Progress struct {
		Status             ProgressStatus `json:"status"`
		TotalBytes         int            `json:"total_bytes,omitempty"`
		TotalBytesEstimate float64        `json:"total_bytes_estimate,omitempty"`
		DownloadedBytes    int            `json:"downloaded_bytes"`
		Filename           string         `json:"filename,omitempty"`
		TmpFilename        string         `json:"tmpfilename,omitempty"`
		FragmentIndex      int            `json:"fragment_index,omitempty"`
		FragmentCount      int            `json:"fragment_count,omitempty"`
		// There are technically other fields, but these are the important ones.
	} `json:"progress"`
	AutoNumber      int `json:"autonumber,omitempty"`
	VideoAutoNumber int `json:"video_autonumber,omitempty"`
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

func (h *progressHandler) parse(raw json.RawMessage) {
	data := &progressData{}

	err := json.Unmarshal(raw, data)
	if err != nil {
		return
	}

	cleanJSON(data)

	update := ProgressUpdate{
		Info:            data.Info,
		Status:          data.Progress.Status,
		TotalBytes:      data.Progress.TotalBytes,
		DownloadedBytes: data.Progress.DownloadedBytes,
		FragmentIndex:   data.Progress.FragmentIndex,
		FragmentCount:   data.Progress.FragmentCount,
		Filename:        data.Progress.Filename,
	}

	if update.TotalBytes == 0 {
		update.TotalBytes = int(data.Progress.TotalBytesEstimate)
	}

	if update.Filename == "" {
		if data.Progress.TmpFilename != "" {
			update.Filename = data.Progress.TmpFilename
		} else if data.Info.Filename != nil && *data.Info.Filename != "" {
			update.Filename = *data.Info.Filename
		}
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
	Info *ExtractedInfo `json:"info"`

	// Status is the current status of the download.
	Status ProgressStatus `json:"status"`
	// TotalBytes is the total number of bytes in the download. If yt-dlp is unable
	// to determine the total bytes, this will be 0.
	TotalBytes int `json:"total_bytes"`
	// DownloadedBytes is the number of bytes that have been downloaded so far.
	DownloadedBytes int `json:"downloaded_bytes"`
	// FragmentIndex is the index of the current fragment being downloaded.
	FragmentIndex int `json:"fragment_index,omitempty"`
	// FragmentCount is the total number of fragments in the download.
	FragmentCount int `json:"fragment_count,omitempty"`

	// Filename is the filename of the video being downloaded, if available. Note that
	// this is not necessarily the same as the destination file, as post-processing
	// may merge multiple files into one.
	Filename string `json:"filename"`

	// Started is the time the download started.
	Started time.Time `json:"started"`
	// Finished is the time the download finished. If the download is still in progress,
	// this will be zero. You can validate with IsZero().
	Finished time.Time `json:"finished,omitempty"`
}

func (p *ProgressUpdate) uuid() string {
	unique := []string{
		p.Filename,
		p.Info.ID,
	}

	if p.Info.PlaylistID != nil {
		unique = append(unique, *p.Info.PlaylistID)
	}

	if p.Info.PlaylistIndex != nil {
		unique = append(unique, strconv.Itoa(*p.Info.PlaylistIndex))
	}

	return strings.Join(unique, ":")
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
		ProgressTemplate(string(progressPrefix) + progressFormat).
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
