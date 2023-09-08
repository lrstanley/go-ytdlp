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

func TestCommand_Simple(t *testing.T) {
	dir := t.TempDir()

	var urls []string

	for _, f := range sampleFiles {
		urls = append(urls, f.url)
	}

	res, err := New().
		Verbose().PrintJson().NoProgress().NoOverwrites().Output(filepath.Join(dir, "%(extractor)s - %(title)s.%(ext)s")).
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

	var hasJSON bool
	for _, l := range res.OutputLogs {
		if l.JSON != nil {
			hasJSON = true
			break
		}
	}

	if !hasJSON {
		t.Fatal("expected at least one log line to be valid JSON due to --print-json")
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

func TestCommand_Version(t *testing.T) {
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

func TestCommand_Unset(t *testing.T) {
	builder := New().NoProgress().Output("test.mp4")

	cmd := builder.buildCommand(context.TODO(), sampleFiles[0].url)

	// Make sure --no-progress is set.
	if !slices.Contains(cmd.Args, "--no-progress") {
		t.Fatal("expected --no-progress flag to be set")
	}

	_ = builder.UnsetProgress()

	cmd = builder.buildCommand(context.TODO(), sampleFiles[0].url)

	// Make sure --no-progress is not set.
	if slices.Contains(cmd.Args, "--no-progress") {
		t.Fatal("expected --no-progress flag to not be set")
	}
}

func TestCommand_Clone(t *testing.T) {
	builder1 := New().NoProgress().Output("test.mp4")

	builder2 := builder1.Clone()

	cmd := builder2.buildCommand(context.TODO(), sampleFiles[0].url)

	// Make sure --no-progress is set.
	if !slices.Contains(cmd.Args, "--no-progress") {
		t.Fatal("expected --no-progress flag to be set")
	}
}

func TestCommand_SetExecutable(t *testing.T) {
	cmd := New().SetExecutable("/usr/bin/test").buildCommand(context.Background(), sampleFiles[0].url)

	if cmd.Path != "/usr/bin/test" {
		t.Fatalf("expected executable to be /usr/bin/test, got %s", cmd.Path)
	}
}

func TestCommand_SetWorkDir(t *testing.T) {
	cmd := New().SetWorkDir("/tmp").buildCommand(context.Background(), sampleFiles[0].url)

	if cmd.Dir != "/tmp" {
		t.Fatalf("expected workdir to be /tmp, got %s", cmd.Dir)
	}
}

func TestCommand_SetEnvVar(t *testing.T) {
	cmd := New().SetEnvVar("TEST", "1").buildCommand(context.Background(), sampleFiles[0].url)

	if cmd.Env[0] != "TEST=1" {
		t.Fatalf("expected env var to be TEST=1, got %s", cmd.Env[0])
	}
}
