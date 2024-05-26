package ytdlp

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

// DownloadStatus represents the status of a download.
type DownloadStatus string

const (
	// StatusDownloading represents the download status when the download is in progress.
	StatusDownloading DownloadStatus = "downloading"
	// StatusMerging represents the download status when the download is finished and the files are
	// being merged.
	StatusMerging DownloadStatus = "merging"
	// StatusFinished represents the download status when the download is finished.
	StatusFinished DownloadStatus = "finished"
)

// DownloadProgress represents the progress of a download.
type DownloadProgress struct {
	Filename   string
	Total      int64
	Downloaded int64
	Speed      int64
	Percent    float64
	Status     DownloadStatus
}

// DownloadProgressFunc is a callback function that is called whenever there is a progress update.
type DownloadProgressFunc = func(DownloadProgress)

// RunDownload runs the command and downloads the video. It returns the download result if successful.
func (c *Command) runDownload(cmd *exec.Cmd, progressFunc DownloadProgressFunc) error {
	if cmd.Err != nil {
		return cmd.Err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	c.applySyscall(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}

	doneStdout, doneStderr := make(chan struct{}), make(chan struct{})
	var stderrBuf bytes.Buffer
	var progress DownloadProgress

	go func() {
		defer close(doneStdout)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			handleProgressUpdate(line, &progress, progressFunc)
		}
	}()

	go func() {
		defer close(doneStderr)
		if _, err := stderrBuf.ReadFrom(stderr); err != nil {
			stderrBuf.WriteString(fmt.Sprintf("failed to read from stderr: %v", err))
		}
	}()

	// Ensure we wait for both stdout/stderr to finish before we wait for the command.
	// This is to ensure we don't miss any data.
	<-doneStdout
	<-doneStderr

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s", err, stderrBuf.String())
	}

	// Command finished successfully, so we can mark the progress as finished.
	progress.Status = StatusFinished
	progressFunc(progress)

	return nil
}

// handleProgressLine parses the progress line and updates the progress struct.
func handleProgressUpdate(line string, progress *DownloadProgress, progressFunc DownloadProgressFunc) {
	if progress.Status == StatusMerging {
		return
	}

	re := regexp.MustCompile(`dl:(.+):(.+):([0-9NA.]+),([0-9NA.]+),([0-9NA.]+),([0-9NA.]+),\s*([\dNA.]+)%`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 8 {
		return
	}

	total, err := strconv.ParseFloat(safeParse(matches[3], matches[4]), 64)
	if err != nil {
		return
	}
	downloaded, err := strconv.ParseFloat(matches[5], 64)
	if err != nil {
		return
	}
	speed, err := strconv.ParseFloat(matches[6], 64)
	if err != nil {
		return
	}
	percent, err := strconv.ParseFloat(matches[7], 64)
	if err != nil {
		return
	}
	filename := matches[1]
	status := matches[2]

	progress.Filename = filename
	progress.Total = int64(total)
	progress.Downloaded = int64(downloaded)
	progress.Speed = int64(speed)
	progress.Percent = percent
	progress.Status = StatusDownloading

	if status == "finished" {
		progress.Percent = 100
		progress.Status = StatusMerging
	}

	progressFunc(*progress)
}

// safeParse returns the primary value if it's not "NA", otherwise it returns the fallback value.
func safeParse(primary, fallback string) string {
	if primary != "NA" {
		return primary
	}
	return fallback
}
