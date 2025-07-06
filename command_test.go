// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
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

func TestMain(m *testing.M) {
	os.Setenv("YTDLP_DEBUG", "true")
	MustInstallAll(context.TODO())
	os.Exit(m.Run())
}

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

	progressUpdates := map[string]ProgressUpdate{}

	res, err := New().
		Verbose().
		PrintJSON().
		NoProgress().
		NoOverwrites().
		Output(filepath.Join(dir, "%(extractor)s - %(title)s.%(ext)s")).
		ProgressFunc(100*time.Millisecond, func(prog ProgressUpdate) {
			progressUpdates[prog.Filename] = prog
		}).
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
		t.Fatal("expected at least one log line to be valid JSON due to one of --print-json/--dump-json/--print '%()j'")
	}

	for _, f := range sampleFiles {
		t.Run(f.name, func(t *testing.T) {
			var stat fs.FileInfo

			fn := filepath.Join(dir, fmt.Sprintf("%s - %s.%s", f.extractor, f.name, f.ext))

			stat, err = os.Stat(fn)
			if err != nil {
				t.Fatal(err)
			}

			if stat.Size() == 0 {
				t.Fatal("file is empty")
			}

			prog, ok := progressUpdates[fn]
			if !ok {
				t.Fatalf("expected progress updates for %s", fn)
			}

			if prog.Finished.IsZero() || prog.Started.IsZero() {
				t.Fatal("expected progress start and finish times to be set")
			}

			if prog.TotalBytes == 0 {
				t.Fatal("expected progress total bytes to be set")
			}
			if prog.DownloadedBytes == 0 {
				t.Fatal("expected progress downloaded bytes to be set")
			}

			if prog.Percent() < 100.0 {
				t.Fatalf("expected progress to be 100%%, got %.2f%%", prog.Percent())
			}

			if prog.Info.URL == nil {
				t.Fatal("expected progress info URL to be set")
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
	builder := New().Progress().NoProgress().Output("test.mp4")

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

	// Progress and NoProgress should conflict, so arg length should be 4 (executable, output, output value, and url).
	if len(cmd.Args) != 4 {
		t.Fatalf("expected arg length to be 4, got %d: %#v", len(cmd.Args), cmd.Args)
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

	if !slices.Contains(cmd.Env, "TEST=1") {
		t.Fatalf("expected env var to be TEST=1, got %v", cmd.Env)
	}
}

func TestCommand_SetFlagConfig_DuplicateFlags(t *testing.T) {
	flagConfig := &FlagConfig{}
	flagConfig.General.IgnoreErrors = ptr(true)
	flagConfig.General.AbortOnError = ptr(true)

	builder := New().SetFlagConfig(flagConfig)

	err := builder.flagConfig.General.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if _, ok := IsMultipleJSONParsingFlagsError(err); !ok {
		t.Fatalf("expected validation error to be a multiple JSON parsing flags error, got %v", err)
	}
}
