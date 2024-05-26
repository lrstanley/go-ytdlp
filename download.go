package ytdlp

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

type DownloadStatus string

const (
	StatusDownloading DownloadStatus = "downloading"
	StatusMerging     DownloadStatus = "merging"
	StatusFinished    DownloadStatus = "finished"
)

type DownloadProgress struct {
	Total      int64
	Downloaded int64
	Speed      int64
	Percent    float64
	Status     DownloadStatus
}

type DownloadProgressFunc = func(DownloadProgress)

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
			fmt.Println("stdout:", line)
			handleProgressLine(line, &progress, progressFunc)
		}
	}()

	go func() {
		defer close(doneStderr)
		if _, err := stderrBuf.ReadFrom(stderr); err != nil {
			fmt.Println("stderr err:", err)
		}
	}()

	// Ensure we wait for both stdout/stderr to finish before we wait for the command.
	// This is to ensure we don't miss any data.
	<-doneStdout
	<-doneStderr

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s", err, stderrBuf.String())
	}

	// Command finished successfully, so we can mark the progress as complete.
	progress.Status = StatusFinished
	progressFunc(progress)

	return nil
}

func handleProgressLine(line string, progress *DownloadProgress, progressFunc DownloadProgressFunc) {
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
	status := matches[2]

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

func safeParse(primary, fallback string) string {
	if primary != "NA" {
		return primary
	}
	return fallback
}
