//go:build e2e

package ytdlp

import (
	"strings"
	"testing"
)

// These URLs are from yt-dlp's GenericIE._TESTS:
// https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/generic.py
func TestExtractedInfoE2E(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		id          string
		extension   string
		title       string
		titlePrefix string
	}{
		{
			name:      "direct webm",
			url:       "https://ftp.nluug.nl/video/nluug/2014-11-20_nj14/zaal-2/5_Lennart_Poettering_-_Systemd.webm",
			id:        "5_Lennart_Poettering_-_Systemd",
			extension: "webm",
			title:     "5_Lennart_Poettering_-_Systemd",
		},
		{
			name:      "m3u8 sample",
			url:       "https://raw.githubusercontent.com/grafov/m3u8/refs/heads/master/sample-playlists/master.m3u8",
			id:        "master",
			extension: "mp4",
			title:     "master",
		},
		{
			name:      "videojs hls",
			url:       "https://gist.githubusercontent.com/bashonly/2aae0862c50f4a4b84f220c315767208/raw/e3380d413749dabbe804c9c2d8fd9a45142475c7/videojs_hls_test.html",
			id:        "videojs_hls_test",
			extension: "mp4",
			title:     "video",
		},
		{
			name:        "live dash",
			url:         "https://livesim2.dashif.org/livesim2/ato_10/testpic_2s/Manifest.mpd",
			id:          "Manifest",
			extension:   "mp4",
			titlePrefix: "Manifest ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, _, err := New().NoUpdate().ExtractInfo(t.Context(), tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if len(info) != 1 {
				t.Fatalf("info has length %d, want 1", len(info))
			}

			got := info[0]
			if got.ID != tt.id {
				t.Errorf("id = %q, want %q", got.ID, tt.id)
			}
			if got.Extension != tt.extension {
				t.Errorf("extension = %q, want %q", got.Extension, tt.extension)
			}
			if got.Title == nil || (tt.title != "" && *got.Title != tt.title) ||
				(tt.titlePrefix != "" && !strings.HasPrefix(*got.Title, tt.titlePrefix)) {
				t.Errorf("title = %v, want %q", got.Title, tt.title)
			}
			if len(got.Formats) == 0 {
				t.Error("formats is empty")
			}
		})
	}
}

// TestExtractedInfoE2EPlaylist covers multiple results returned by an XSPF
// playlist, which is not exercised by the single-video cases above.
func TestExtractedInfoE2EPlaylist(t *testing.T) {
	const playlistURL = "https://shellac-archive.ch/repository/xspf/22-AL0019Z.xspf"

	info, _, err := New().NoUpdate().ExtractInfo(t.Context(), playlistURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(info) < 2 {
		t.Fatalf("info has length %d, want multiple results", len(info))
	}

	var entry *ExtractedInfo
	for _, candidate := range info {
		if candidate.ID == "22-AL0019Z" {
			entry = candidate
			break
		}
	}
	if entry == nil {
		t.Fatal("playlist entry not found")
	}
	if entry.ID != "22-AL0019Z" {
		t.Errorf("entry id = %q, want %q", entry.ID, "22-AL0019Z")
	}
	if entry.Extension != "mp3" {
		t.Errorf("entry extension = %q, want %q", entry.Extension, "mp3")
	}
	if entry.Title == nil || *entry.Title != "Concerto in B Flat Major (Brahms, Op. 83)" {
		t.Errorf("entry title = %v, want %q", entry.Title, "Concerto in B Flat Major (Brahms, Op. 83)")
	}
	if len(entry.Formats) == 0 {
		t.Error("entry formats is empty")
	}
}

func TestExtractedInfoE2EYouTube(t *testing.T) {
	const videoURL = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	info, _, err := New().NoUpdate().ExtractInfo(t.Context(), videoURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(info) != 1 {
		t.Fatalf("info has length %d, want 1", len(info))
	}

	got := info[0]
	if got.ID != "dQw4w9WgXcQ" {
		t.Errorf("id = %q, want %q", got.ID, "dQw4w9WgXcQ")
	}
	if got.Title == nil || !strings.HasPrefix(*got.Title, "Rick Astley - Never Gonna Give You Up") {
		t.Errorf("title = %v, want Rick Astley - Never Gonna Give You Up", got.Title)
	}
	if got.Uploader == nil || *got.Uploader != "Rick Astley" {
		t.Errorf("uploader = %v, want %q", got.Uploader, "Rick Astley")
	}
	if got.UploaderID == nil || *got.UploaderID != "@RickAstleyYT" {
		t.Errorf("uploader id = %v, want %q", got.UploaderID, "@RickAstleyYT")
	}
	if got.UploaderURL == nil || *got.UploaderURL != "https://www.youtube.com/@RickAstleyYT" {
		t.Errorf("uploader url = %v, want %q", got.UploaderURL, "https://www.youtube.com/@RickAstleyYT")
	}
	if got.Channel == nil || *got.Channel != "Rick Astley" {
		t.Errorf("channel = %v, want %q", got.Channel, "Rick Astley")
	}
	if got.ChannelID == nil || *got.ChannelID != "UCuAXFkgsw1L7xaCfnd5JJOw" {
		t.Errorf("channel id = %v, want %q", got.ChannelID, "UCuAXFkgsw1L7xaCfnd5JJOw")
	}
	if got.ChannelURL == nil || *got.ChannelURL != "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw" {
		t.Errorf("channel url = %v, want %q", got.ChannelURL, "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw")
	}
	if got.UploadDate == nil || *got.UploadDate != "20091025" {
		t.Errorf("upload date = %v, want %q", got.UploadDate, "20091025")
	}
	if got.Duration == nil || *got.Duration != 213 {
		t.Errorf("duration = %v, want 213", got.Duration)
	}
	if got.AgeLimit == nil || *got.AgeLimit != 0 {
		t.Errorf("age limit = %v, want 0", got.AgeLimit)
	}
	if got.Availability == nil || *got.Availability != ExtractedAvailabilityPublic {
		t.Errorf("availability = %v, want %q", got.Availability, ExtractedAvailabilityPublic)
	}
	if got.LiveStatus == nil || *got.LiveStatus != ExtractedLiveStatusNotLive {
		t.Errorf("live status = %v, want %q", got.LiveStatus, ExtractedLiveStatusNotLive)
	}
	if got.IsLive == nil || *got.IsLive {
		t.Errorf("is live = %v, want false", got.IsLive)
	}
	if got.WasLive == nil || *got.WasLive {
		t.Errorf("was live = %v, want false", got.WasLive)
	}
	if got.ViewCount == nil || *got.ViewCount <= 0 {
		t.Errorf("view count = %v, want a positive value", got.ViewCount)
	}
	if got.LikeCount == nil || *got.LikeCount <= 0 {
		t.Errorf("like count = %v, want a positive value", got.LikeCount)
	}
	if got.WebpageURL == nil || *got.WebpageURL != videoURL {
		t.Errorf("webpage url = %v, want %q", got.WebpageURL, videoURL)
	}
	if got.Thumbnail == nil || !strings.HasPrefix(*got.Thumbnail, "https://i.ytimg.com/") {
		t.Errorf("thumbnail = %v, want an i.ytimg.com URL", got.Thumbnail)
	}
	if len(got.Categories) == 0 || got.Categories[0] != "Music" {
		t.Errorf("categories = %v, want Music", got.Categories)
	}
	if len(got.Tags) == 0 {
		t.Error("tags is empty")
	}
	if len(got.Subtitles) == 0 {
		t.Error("subtitles is empty")
	}
	if len(got.AutomaticCaptions) == 0 {
		t.Error("automatic captions is empty")
	}
	if len(got.Thumbnails) == 0 {
		t.Error("thumbnails is empty")
	}
	if len(got.Formats) == 0 {
		t.Error("formats is empty")
	}

	var hasAudio, hasVideo bool
	for _, format := range got.Formats {
		if format.ACodec != nil && *format.ACodec != "none" {
			hasAudio = true
		}
		if format.VCodec != nil && *format.VCodec != "none" {
			hasVideo = true
		}
	}
	if !hasAudio {
		t.Error("formats have no audio-only or muxed format")
	}
	if !hasVideo {
		t.Error("formats have no video-only or muxed format")
	}
}
