// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json/jsontext"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestProgressHandler_parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		raw           string
		wantStatus    ProgressStatus
		wantProcessor string
	}{
		{
			name:       "download",
			raw:        `{"info":{"id":"abc"},"progress":{"status":"downloading","filename":"a.mp4"}}`,
			wantStatus: ProgressStatusDownloading,
		},
		{
			name:       "download finished",
			raw:        `{"info":{"id":"abc"},"progress":{"status":"finished","filename":"a.mp4"}}`,
			wantStatus: ProgressStatusFinished,
		},
		{
			name:          "postprocess finished",
			raw:           `{"info":{"id":"abc","filename":"a.mp4"},"progress":{"status":"finished","postprocessor":"Merger"}}`,
			wantStatus:    ProgressStatusPostProcessing,
			wantProcessor: "Merger",
		},
		{
			name:          "postprocess error",
			raw:           `{"info":{"id":"abc"},"progress":{"status":"error","postprocessor":"Merger"}}`,
			wantStatus:    ProgressStatusError,
			wantProcessor: "Merger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got ProgressUpdate
			h := newProgressHandler(func(update ProgressUpdate) {
				got = update
			})
			h.parse(jsontext.Value(tt.raw))

			if got.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", got.Status, tt.wantStatus)
			}
			if got.PostProcessor != tt.wantProcessor {
				t.Errorf("post processor = %q, want %q", got.PostProcessor, tt.wantProcessor)
			}
		})
	}
}

func TestProgressUpdate_Key(t *testing.T) {
	t.Parallel()

	info := &ExtractedInfo{ID: "abc"}
	video := ProgressUpdate{Filename: "a.f137.mp4", Info: info}
	audio := ProgressUpdate{Filename: "a.f140.m4a", Info: info}
	merge := ProgressUpdate{Filename: "a.mp4", Info: info, PostProcessor: "Merger"}

	if video.Key() == audio.Key() {
		t.Errorf("video and audio keys are equal: %q", video.Key())
	}
	if audio.Key() == merge.Key() {
		t.Errorf("audio and merge keys are equal: %q", audio.Key())
	}
}

func TestProgressFunc_templates(t *testing.T) {
	t.Parallel()

	cfg := New().
		ProgressFunc(100*time.Millisecond, func(ProgressUpdate) {}).
		GetFlagConfig()

	want := []string{
		"download:" + string(progressPrefix) + progressFormat,
		"postprocess:" + string(progressPrefix) + progressFormat,
	}
	if !slices.Equal(cfg.VerbositySimulation.ProgressTemplate, want) {
		t.Errorf("progress templates = %v, want %v", cfg.VerbositySimulation.ProgressTemplate, want)
	}
}

func TestCommand_ProgressPostProcess(t *testing.T) {
	t.Parallel()

	server := newMockServer(t, "testdata/sample-1.mp4")
	dir := t.TempDir()

	var mu sync.Mutex
	var updates []ProgressUpdate

	result, err := New().
		NoUpdate().
		ForceOverwrites().
		RecodeVideo("mkv").
		Output(filepath.Join(dir, "%(extractor)s - %(title)s.%(ext)s")).
		ProgressFunc(100*time.Millisecond, func(update ProgressUpdate) {
			mu.Lock()
			updates = append(updates, update)
			mu.Unlock()
		}).
		Run(t.Context(), server.fileURL)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}

	mu.Lock()
	defer mu.Unlock()

	var sawDownload, sawPost bool
	for _, update := range updates {
		if update.PostProcessor == "" {
			sawDownload = true
			continue
		}
		sawPost = true
		if update.Status != ProgressStatusPostProcessing {
			t.Errorf("post-processing status = %v, want %v", update.Status, ProgressStatusPostProcessing)
		}
	}

	if !sawDownload {
		t.Error("expected download progress")
	}
	if !sawPost {
		t.Error("expected post-processing progress")
	}
}
