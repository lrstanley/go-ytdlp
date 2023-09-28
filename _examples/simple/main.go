// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)

	dl := ytdlp.New().
		PrintJSON().
		NoProgress().
		FormatSort("res,ext:mp4:m4a").
		RecodeVideo("mp4").
		NoPlaylist().
		NoOverwrites().
		Continue().
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

	log.Println("wrote results to results.json")
}
