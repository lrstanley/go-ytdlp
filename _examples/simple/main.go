// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	// Use the following env var if you want to debug the download process for yt-dlp, ffmpeg, and ffprobe,
	// as well as print out associated yt-dlp commands.
	// os.Setenv("YTDLP_DEBUG", "true")

	// If yt-dlp/ffmpeg/ffprobe isn't installed yet, download and cache the binaries for further use.
	// Note that the download/installation of ffmpeg/ffprobe is only supported on a handful of platforms,
	// and so it is still recommended to install ffmpeg/ffprobe via other means.
	ytdlp.MustInstallAll(context.TODO())

	dl := ytdlp.New().
		PrintJSON().
		NoProgress().
		FormatSort("res,ext:mp4:m4a").
		RecodeVideo("mp4").
		NoPlaylist().
		NoOverwrites().
		Continue().
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			fmt.Printf( //nolint:forbidigo
				"%s @ %s [eta: %s] :: %s\n",
				prog.Status,
				prog.PercentString(),
				prog.ETA(),
				prog.Filename,
			)
		}).
		Output("%(extractor)s - %(title)s.%(ext)s")

	r, err := dl.Run(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		panic(err)
	}

	f, err := os.Create("results.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")

	if err = enc.Encode(r); err != nil {
		panic(err)
	}

	slog.Info("wrote results to results.json")
}
