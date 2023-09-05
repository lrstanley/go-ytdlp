// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:forbidigo
package ytdlp

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

type testSampleFile struct {
	url       string
	name      string
	ext       string
	extractor string
}

var sampleFiles = []testSampleFile{
	{url: "https://cdn.liam.sh/github/go-ytdlp/sample-1.mp4", name: "sample-1", ext: "mp4", extractor: "generic"},
	{url: "https://cdn.liam.sh/github/go-ytdlp/sample-2.mp4", name: "sample-2", ext: "mp4", extractor: "generic"},
	{url: "https://cdn.liam.sh/github/go-ytdlp/sample-3.mp4", name: "sample-3", ext: "mp4", extractor: "generic"},
	{url: "https://cdn.liam.sh/github/go-ytdlp/sample-4.mpg", name: "sample-4", ext: "mpg", extractor: "generic"},
}

func TestCommandSimple(t *testing.T) {
	dir := t.TempDir()

	var urls []string

	for _, f := range sampleFiles {
		urls = append(urls, f.url)
	}

	res, err := New().
		Verbose().NoProgress().NoOverwrites().Output(filepath.Join(dir, "%(extractor)s - %(title)s.%(ext)s")).
		Run(context.Background(), urls...)
	if err != nil {
		t.Fatal(err)
	}

	if res == nil {
		t.Fatal("res is nil")
	}

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}

	if !slices.Contains(res.Args, "--verbose") {
		t.Fatal("expected --verbose flag to be set")
	}

	for _, f := range sampleFiles {
		t.Run(f.name, func(t *testing.T) {
			var stat fs.FileInfo

			stat, err = os.Stat(filepath.Join(dir, fmt.Sprintf("%s - %s.%s", f.extractor, f.name, f.ext)))
			if err != nil {
				t.Fatal(err)
			}

			if stat.Size() == 0 {
				t.Fatal("file is empty")
			}
		})
	}
}

func TestCommandVersion(t *testing.T) {
	res, err := New().Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if res == nil {
		t.Fatal("res is nil")
	}

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}

	_, err = time.Parse("2006.01.02", res.Stdout)
	if err != nil {
		t.Fatalf("failed to parse version: %v", err)
	}
}
