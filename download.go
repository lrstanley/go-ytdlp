package ytdlp

import (
	"encoding/json"
	"strconv"
	"strings"
)

// DownloadProgressFunc is a callback function that is called whenever there is a progress update.
type DownloadProgressFunc = func(DownloadProgress)

// DownloadStatus represents the status of a download.
type DownloadStatus string

const (
	// StatusPreProcessing represents the download status when the download is pre-processing.
	StatusStarting DownloadStatus = "starting"
	// StatusDownloading represents the download status when the download is in progress.
	StatusDownloading DownloadStatus = "downloading"
	// StatusPostProcessing represents the download status when the download is post-processing.
	StatusPostProcessing DownloadStatus = "post_processing"
	// StatusFinished represents the download status when the download is finished.
	StatusFinished DownloadStatus = "finished"
)

// DownloadProgress represents the progress of a download.
type DownloadProgress struct {
	ID            string
	PlaylistID    string
	Title         string
	Total         int64
	Downloaded    int64
	Speed         int64
	Percent       float64
	Status        DownloadStatus
	PlaylistCount int
	PlaylistIndex int
}

func (p *DownloadProgress) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID                 string `json:"id"`
		PlaylistID         string `json:"playlist_id"`
		Title              string `json:"title"`
		Status             string `json:"status"`
		TotalBytes         string `json:"total_bytes"`
		TotalBytesEstimate string `json:"total_bytes_estimate"`
		DownloadedBytes    string `json:"downloaded_bytes"`
		Speed              string `json:"speed"`
		PercentStr         string `json:"percent_str"`
		PlaylistCount      string `json:"playlist_count"`
		PlaylistIndex      string `json:"playlist_index"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.ID = raw.ID
	p.PlaylistID = safeParse(raw.PlaylistID, "")
	p.Title = raw.Title

	total := safeParse(raw.TotalBytes, raw.TotalBytesEstimate)
	totalFloat, _ := strconv.ParseFloat(total, 64)
	p.Total = int64(totalFloat)

	downloaded := safeParse(raw.DownloadedBytes, "0")
	downloadedFloat, _ := strconv.ParseFloat(downloaded, 64)
	p.Downloaded = int64(downloadedFloat)

	speed := safeParse(raw.Speed, "0")
	speedFloat, _ := strconv.ParseFloat(speed, 64)
	p.Speed = int64(speedFloat)

	percent := strings.TrimSpace(strings.TrimSuffix(raw.PercentStr, "%"))
	percentFloat, _ := strconv.ParseFloat(percent, 64)
	p.Percent = percentFloat

	playlistCount := safeParse(raw.PlaylistCount, "0")
	playlistCountInt, _ := strconv.Atoi(playlistCount)
	p.PlaylistCount = playlistCountInt

	playListIndex := safeParse(raw.PlaylistIndex, "0")
	playListIndexInt, _ := strconv.Atoi(playListIndex)
	p.PlaylistIndex = playListIndexInt

	p.Status = StatusDownloading

	return nil
}

func (p *DownloadProgress) IsPlaylist() bool {
	return p.PlaylistCount > 0
}

// progressStatus represents the status of a progress event.
type progressStatus string

const (
	progressPreProcessing      progressStatus = "pre_processing"
	progressStarting           progressStatus = "starting"
	progressDownloading        progressStatus = "downloading"
	progressFinished           progressStatus = "finished"
	progressPostProcessing     progressStatus = "post_processing"
	progressVideoDownloaded    progressStatus = "video_downloaded"
	progressPlaylistDownloaded progressStatus = "playlist_downloaded"
)

// progressEvent represents a download progress event.
type progressEvent struct {
	ID         string         `json:"id"`
	PlaylistID string         `json:"playlist_id"`
	Title      string         `json:"title"`
	Status     progressStatus `json:"status"`
}

// safeParse returns the primary value if it's not "NA", otherwise it returns the fallback value.
func safeParse(primary, fallback string) string {
	if primary != "NA" {
		return primary
	}
	return fallback
}
