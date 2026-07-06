// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockServer struct {
	*httptest.Server

	fileURL string
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

	require.Len(t, info, 1, "expected 1 extracted info")
	require.NotNil(t, info[0].FormatID, "expected format id to be set")
	assert.Equal(t, "mp4", *info[0].FormatID, "expected format id to be mp4")

	require.NotNil(t, info[0].Protocol, "expected protocol to be set")
	assert.Equal(t, "http", *info[0].Protocol, "expected protocol to be http")

	require.NotNil(t, info[0].HTTPHeaders, "expected http headers to be set")
	assert.Contains(t, info[0].HTTPHeaders["User-Agent"], "Mozilla", "expected User-Agent header to be set and contain Mozilla")

	assert.Equal(t, "sample-1", info[0].ID, "expected id to be set")

	require.NotNil(t, info[0].Title, "expected title to be set")
	assert.Equal(t, "sample-1", *info[0].Title, "expected title to be set")

	require.Len(t, info[0].Formats, 1, "expected 1 format")
	require.NotNil(t, info[0].Formats[0].Extension, "expected format extension to be set")
	assert.Equal(t, "mp4", *info[0].Formats[0].Extension, "expected format extension to be mp4")

	require.NotNil(t, info[0].URL, "expected url to be set")
	assert.Equal(t, server.fileURL, *info[0].URL, "expected url to be set")
	require.NotNil(t, info[0].WebpageURL, "expected webpage url to be set")
	assert.Equal(t, server.fileURL, *info[0].WebpageURL, "expected webpage url to be set")

	require.NotNil(t, info[0].Filename, "expected filename to be set")
	assert.FileExists(t, *info[0].Filename, "expected file to exist")

	require.NotNil(t, info[0].Timestamp, "expected timestamp to be set")
	assert.Positive(t, *info[0].Timestamp, "expected timestamp to be set")

	require.NotNil(t, info[0].UploadDate, "expected upload date to be set")
	assert.Positive(t, *info[0].UploadDate, "expected upload date to be set")

	require.NotNil(t, info[0].Extractor, "expected extractor to be set")
	assert.Equal(t, "generic", *info[0].Extractor, "expected extractor to be generic")

	require.NotNil(t, info[0].ExtractorKey, "expected extractor key to be set")
	assert.Equal(t, "Generic", *info[0].ExtractorKey, "expected extractor key to be generic")
}

func TestParseExtractedInfo_requestedSubtitles(t *testing.T) {
	// yt-dlp's process_subtitles selects a single subtitle format per language,
	// so requested_subtitles holds one object per language, while subtitles and
	// automatic_captions hold a list of formats per language.
	raw := json.RawMessage(`{
		"id": "sample-1",
		"subtitles": {"en": [{"ext": "vtt", "url": "https://example.com/en.vtt"}, {"ext": "srt", "url": "https://example.com/en.srt"}]},
		"automatic_captions": {"en": [{"ext": "vtt", "url": "https://example.com/auto-en.vtt"}]},
		"requested_subtitles": {"en": {"ext": "srt", "url": "https://example.com/en.srt", "name": "English"}}
	}`)

	info, err := ParseExtractedInfo(&raw)
	require.NoError(t, err, "expected requested_subtitles to unmarshal")

	require.Contains(t, info.RequestedSubtitles, "en", "expected en requested subtitle")
	assert.Equal(t, "https://example.com/en.srt", info.RequestedSubtitles["en"].URL, "expected en requested subtitle url")

	require.Len(t, info.Subtitles["en"], 2, "expected 2 en subtitle formats")
	require.Len(t, info.AutomaticCaptions["en"], 1, "expected 1 en automatic caption format")
}
