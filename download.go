package ytdlp

import (
	"fmt"

	"github.com/lrstanley/go-ytdlp/template"
)

const progressLinePrefix = "dl:"

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
	// ID represents the video ID.
	ID string `ytdlp:"info.id"`
	// PlaylistID represents the playlist id if the download is part of a playlist.
	PlaylistID string `ytdlp:"info.playlist_id"`
	// Title represents the title of the video.
	Title string `ytdlp:"info.title"`
	// TotalBytes represents the size of the video in bytes.
	// if the ytdlp total_bytes doesn't exist, it will then fallback to the total_bytes_estimate.
	TotalBytes int64 `ytdlp:"progress.total_bytes"`
	// TotalBytesEstimated represents the estimated size of the video in bytes.
	TotalBytesEstimated int64 `ytdlp:"progress.total_bytes_estimate"`
	// DownloadedBytes represents the number of bytes downloaded.
	DownloadedBytes int64 `ytdlp:"progress.downloaded_bytes"`
	// Speed represents the download speed in bytes per second.
	Speed int64 `ytdlp:"progress.speed"`
	// Percent represents the download progress as a percentage.
	Percent float64 `ytdlp:"progress._percent_str,formatter=percentToNumber"`
	// Status represents the status of the download.
	Status DownloadStatus `ytdlp:"progress.status"`
	// PlaylistCount represents the total number of videos in the playlist.
	PlaylistCount int `ytdlp:"info.playlist_count"`
	// PlaylistIndex represents the index of the video in the playlist.
	PlaylistIndex int `ytdlp:"info.playlist_index"`
}

func GetDownloadProgressTemplate() (string, error) {
	var downloadProgress DownloadProgress
	templ, err := template.MarshalTemplate(downloadProgress)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %w", err)
	}

	return progressLinePrefix + string(templ), nil
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
	ID         string `ytdlp:"id"`
	PlaylistID string `ytdlp:"playlist_id"`
	Title      string `ytdlp:"title"`
	Status     progressStatus
}

func GetProgressPreProcessTemplate() (string, error) {
	event := progressEvent{Status: progressPreProcessing}
	templ, err := template.MarshalTemplate(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %w", err)
	}

	return fmt.Sprintf("pre_process:%s%s", progressLinePrefix, string(templ)), nil
}

func GetProgressBeforeDownloadTemplate() (string, error) {
	event := progressEvent{Status: progressStarting}
	templ, err := template.MarshalTemplate(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %w", err)
	}

	return fmt.Sprintf("before_dl:%s%s", progressLinePrefix, string(templ)), nil
}

func GetProgressPostProcessTemplate() (string, error) {
	event := struct {
		ID         string `ytdlp:"id"`
		PlaylistID string `ytdlp:"playlist_id"`
		Status     progressStatus
	}{Status: progressPostProcessing}
	templ, err := template.MarshalTemplate(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %w", err)
	}

	return fmt.Sprintf("post_process:%s%s", progressLinePrefix, string(templ)), nil
}

func GetProgressVideoDownloadedTemplate() (string, error) {
	event := struct {
		ID         string `ytdlp:"id"`
		PlaylistID string `ytdlp:"playlist_id"`
		Status     progressStatus
	}{Status: progressVideoDownloaded}
	templ, err := template.MarshalTemplate(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %w", err)
	}

	return fmt.Sprintf("after_video:%s%s", progressLinePrefix, string(templ)), nil
}

func GetProgressPlaylistDownloadedTemplate() (string, error) {
	event := struct {
		ID     string `ytdlp:"id"`
		Status progressStatus
	}{Status: progressPlaylistDownloaded}
	templ, err := template.MarshalTemplate(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %w", err)
	}

	return fmt.Sprintf("playlist:%s%s", progressLinePrefix, string(templ)), nil
}
