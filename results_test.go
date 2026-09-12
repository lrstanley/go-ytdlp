// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"context"
	"encoding/json/jsontext"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockServer struct {
	*httptest.Server

	fileURL string
}

func TestCleanJSON(t *testing.T) {
	t.Parallel()

	type nested struct {
		Name  string
		Empty *string
	}
	type document struct {
		Name     string
		Title    *string
		Uploader *string
		Nested   *nested
		Items    []*nested
	}

	value := &document{
		Name:     "none",
		Title:    new("none"),
		Uploader: new(""),
		Nested:   &nested{Name: "none", Empty: new("")},
		Items:    []*nested{{Name: "none", Empty: new("none")}},
	}

	cleanJSON(value)

	if value.Name != "" {
		t.Errorf("name = %q, want empty", value.Name)
	}
	if value.Title == nil {
		t.Fatal("title is nil")
	}
	if *value.Title != "" {
		t.Errorf("title = %q, want empty", *value.Title)
	}
	if value.Uploader != nil {
		t.Errorf("uploader = %q, want nil", *value.Uploader)
	}
	if value.Nested == nil {
		t.Fatal("nested is nil")
	}
	if value.Nested.Name != "" {
		t.Errorf("nested name = %q, want empty", value.Nested.Name)
	}
	if value.Nested.Empty != nil {
		t.Errorf("nested empty = %q, want nil", *value.Nested.Empty)
	}
	if len(value.Items) != 1 {
		t.Fatalf("items has length %d, want 1", len(value.Items))
	}
	if value.Items[0].Name != "" {
		t.Errorf("item name = %q, want empty", value.Items[0].Name)
	}
	if value.Items[0].Empty != nil {
		t.Errorf("item empty = %q, want nil", *value.Items[0].Empty)
	}
}

func newMockServer(t *testing.T, fileName string) *mockServer {
	t.Helper()

	base := filepath.Base(fileName)

	// TODO: potentially replace with FileServer + Go 1.24 os.Root.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, base) {
			http.ServeFile(w, r, fileName)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	t.Cleanup(server.Close)

	return &mockServer{
		Server:  server,
		fileURL: server.URL + "/" + base,
	}
}

func TestExtractedInfo(t *testing.T) {
	server := newMockServer(t, "testdata/sample-1.mp4")

	dir := t.TempDir()

	result, err := New().
		ForceOverwrites().
		Output(filepath.Join(dir, "%(extractor)s - %(title)s.%(ext)s")).
		PrintJSON().
		Run(context.TODO(), server.fileURL)
	if err != nil {
		t.Fatal(err)
		return
	}

	info, err := result.GetExtractedInfo()
	if err != nil {
		t.Fatal(err)
	}

	if len(info) != 1 {
		t.Fatalf("info has length %d, want 1", len(info))
	}
	if info[0].FormatID == nil {
		t.Fatal("expected format id to be set")
	}
	if got := *info[0].FormatID; got != "mp4" {
		t.Errorf("format id = %q, want mp4", got)
	}

	if info[0].Protocol == nil {
		t.Fatal("expected protocol to be set")
	}
	if got := *info[0].Protocol; got != "http" {
		t.Errorf("protocol = %q, want http", got)
	}

	if info[0].HTTPHeaders == nil {
		t.Fatal("expected http headers to be set")
	}
	if !strings.Contains(info[0].HTTPHeaders["User-Agent"], "Mozilla") {
		t.Errorf("expected User-Agent header to contain Mozilla, got %q", info[0].HTTPHeaders["User-Agent"])
	}

	if info[0].ID != "sample-1" {
		t.Errorf("id = %q, want sample-1", info[0].ID)
	}

	if info[0].Title == nil {
		t.Fatal("expected title to be set")
	}
	if got := *info[0].Title; got != "sample-1" {
		t.Errorf("title = %q, want sample-1", got)
	}

	if len(info[0].Formats) != 1 {
		t.Fatalf("formats has length %d, want 1", len(info[0].Formats))
	}
	if info[0].Formats[0].Extension == nil {
		t.Fatal("expected format extension to be set")
	}
	if got := *info[0].Formats[0].Extension; got != "mp4" {
		t.Errorf("format extension = %q, want mp4", got)
	}

	if info[0].URL == nil {
		t.Fatal("expected url to be set")
	}
	if got := *info[0].URL; got != server.fileURL {
		t.Errorf("url = %q, want %q", got, server.fileURL)
	}
	if info[0].WebpageURL == nil {
		t.Fatal("expected webpage url to be set")
	}
	if got := *info[0].WebpageURL; got != server.fileURL {
		t.Errorf("webpage url = %q, want %q", got, server.fileURL)
	}

	if info[0].Filename == nil {
		t.Fatal("expected filename to be set")
	}
	if _, statErr := os.Stat(*info[0].Filename); statErr != nil {
		t.Fatalf("expected file to exist: %v", statErr)
	}

	if info[0].Timestamp == nil {
		t.Fatal("expected timestamp to be set")
	}
	if *info[0].Timestamp <= 0 {
		t.Errorf("timestamp = %v, want positive", *info[0].Timestamp)
	}

	if info[0].UploadDate == nil {
		t.Fatal("expected upload date to be set")
	}
	if *info[0].UploadDate == "" {
		t.Error("upload date is empty")
	}

	if info[0].Extractor == nil {
		t.Fatal("expected extractor to be set")
	}
	if got := *info[0].Extractor; got != "generic" {
		t.Errorf("extractor = %q, want generic", got)
	}

	if info[0].ExtractorKey == nil {
		t.Fatal("expected extractor key to be set")
	}
	if got := *info[0].ExtractorKey; got != "Generic" {
		t.Errorf("extractor key = %q, want Generic", got)
	}
}

func TestGetExtractedInfo_dumpJSONFlags(t *testing.T) {
	server := newMockServer(t, "testdata/sample-1.mp4")

	tests := []struct {
		name string
		cmd  *Command
	}{
		{name: "DumpJSON", cmd: New().DumpJSON()},
		{name: "DumpSingleJSON", cmd: New().DumpSingleJSON()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.cmd.NoUpdate().Run(context.TODO(), server.fileURL)
			if err != nil {
				t.Fatal(err)
			}
			if result.Stdout == "" {
				t.Fatal("expected JSON on stdout")
			}
			if !jsontext.Value(result.Stdout).IsValid() {
				t.Fatal("expected stdout to be valid JSON")
			}

			var jsonLogs int
			for _, l := range result.OutputLogs {
				if l.JSON != nil {
					jsonLogs++
				}
			}
			if jsonLogs <= 0 {
				t.Fatal("expected at least one OutputLog with parsed JSON")
			}

			info, err := result.GetExtractedInfo()
			if err != nil {
				t.Fatal(err)
			}
			if len(info) != 1 {
				t.Fatalf("info has length %d, want 1", len(info))
			}
			if info[0].ID != "sample-1" {
				t.Errorf("id = %q, want sample-1", info[0].ID)
			}
			if info[0].Title == nil {
				t.Fatal("title is nil")
			}
			if got := *info[0].Title; got != "sample-1" {
				t.Errorf("title = %q, want sample-1", got)
			}
		})
	}
}

func TestParseExtractedInfo_requestedSubtitles(t *testing.T) {
	// yt-dlp's process_subtitles selects a single subtitle format per language,
	// so requested_subtitles holds one object per language, while subtitles and
	// automatic_captions hold a list of formats per language.
	raw := jsontext.Value(`{
		"id": "sample-1",
		"subtitles": {"en": [{"ext": "vtt", "url": "https://example.com/en.vtt"}, {"ext": "srt", "url": "https://example.com/en.srt"}]},
		"automatic_captions": {"en": [{"ext": "vtt", "url": "https://example.com/auto-en.vtt"}]},
		"requested_subtitles": {"en": {"ext": "srt", "url": "https://example.com/en.srt", "name": "English"}}
	}`)

	info, err := ParseExtractedInfo(&raw)
	if err != nil {
		t.Fatal(err)
	}

	subtitle, ok := info.RequestedSubtitles["en"]
	if !ok {
		t.Fatal("expected en requested subtitle")
	}
	if got := subtitle.URL; got != "https://example.com/en.srt" {
		t.Errorf("requested subtitle URL = %q, want https://example.com/en.srt", got)
	}

	if got := len(info.Subtitles["en"]); got != 2 {
		t.Errorf("en subtitles has length %d, want 2", got)
	}
	if got := len(info.AutomaticCaptions["en"]); got != 1 {
		t.Errorf("en automatic captions has length %d, want 1", got)
	}
}
