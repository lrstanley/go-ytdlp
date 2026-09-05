// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			h.parse(json.RawMessage(tt.raw))

			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantProcessor, got.PostProcessor)
		})
	}
}

func TestProgressUpdate_Key(t *testing.T) {
	t.Parallel()

	info := &ExtractedInfo{ID: "abc"}
	video := ProgressUpdate{Filename: "a.f137.mp4", Info: info}
	audio := ProgressUpdate{Filename: "a.f140.m4a", Info: info}
	merge := ProgressUpdate{Filename: "a.mp4", Info: info, PostProcessor: "Merger"}

	assert.NotEqual(t, video.Key(), audio.Key())
	assert.NotEqual(t, audio.Key(), merge.Key())
}

func TestProgressFunc_templates(t *testing.T) {
	t.Parallel()

	cfg := New().
		ProgressFunc(100*time.Millisecond, func(ProgressUpdate) {}).
		GetFlagConfig()

	assert.Equal(t, []string{
		"download:" + string(progressPrefix) + progressFormat,
		"postprocess:" + string(progressPrefix) + progressFormat,
	}, cfg.VerbositySimulation.ProgressTemplate)
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
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	mu.Lock()
	defer mu.Unlock()

	var sawDownload, sawPost bool
	for _, update := range updates {
		if update.PostProcessor == "" {
			sawDownload = true
			continue
		}
		sawPost = true
		assert.Equal(t, ProgressStatusPostProcessing, update.Status)
	}

	assert.True(t, sawDownload, "expected download progress")
	assert.True(t, sawPost, "expected post-processing progress")
}
