package main

import (
	"context"
	"log"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)

	dl := ytdlp.New().
		Continue().
		Format("ba").
		ExtractAudio().
		AudioFormat("mp3").
		Output("%(extractor)s - %(title)s.%(ext)s")

	err := dl.Download(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ", func(p ytdlp.DownloadProgress) {
		log.Printf("[%s]: downloaded %d/%d bytes (%.2f%%) at %d bytes/s", p.Status, p.Downloaded, p.Total, p.Percent, p.Speed)
	})
	if err != nil {
		panic(err)
	}

	log.Println("Downloaded!")
}
