package main

import (
	"context"
	"fmt"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)

	r, err := ytdlp.New().
		Continue().
		Format("ba").
		ExtractAudio().
		AudioFormat("mp3").
		Output("%(extractor)s - %(title)s.%(ext)s").
		SetProgressFn(func(p ytdlp.DownloadProgress) {
			if p.IsPlaylist() {
				fmt.Printf("\r%s: %.2f%% (%s) [%d/%d] [%d/%d]", p.Title, p.Percent, p.Status, p.Downloaded, p.Total, p.PlaylistIndex, p.PlaylistCount)
			} else {
				fmt.Printf("\r%s: %.2f%% (%s) [%d/%d]", p.Title, p.Percent, p.Status, p.Downloaded, p.Total)
			}
		}).
		Run(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("\rDownloaded %d items", len(r.Downloads))
}
