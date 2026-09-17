// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in the
// LICENSE file.

package ytdlp

import "testing"

func TestFlattenExtractedInfo(t *testing.T) {
	t.Parallel()

	video := &ExtractedInfo{ID: "v1", Type: ExtractedTypeVideo}
	leafURL := &ExtractedInfo{ID: "u1", Type: ExtractedTypeURL}
	got := FlattenExtractedInfo([]*ExtractedInfo{
		nil,
		video,
		{
			Type: ExtractedTypePlaylist,
			Entries: []*ExtractedInfo{
				nil,
				{ID: "p1", Type: ExtractedTypeVideo},
				{
					Type:    ExtractedTypeMultiVideo,
					Entries: []*ExtractedInfo{{ID: "m1", Type: ExtractedTypeURLTransparent}},
				},
			},
		},
		leafURL,
		{ID: "empty"},
	})

	want := []string{"v1", "p1", "m1", "u1", "empty"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), idsOf(got))
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Errorf("got[%d].ID = %q, want %q", i, got[i].ID, id)
		}
	}
	if n := len(FlattenExtractedInfo(nil)); n != 0 {
		t.Fatalf("FlattenExtractedInfo(nil) len = %d", n)
	}
}

func TestExtractedInfoBestThumbnail(t *testing.T) {
	t.Parallel()

	if (*ExtractedInfo)(nil).BestThumbnail() != nil {
		t.Fatal("nil receiver")
	}
	if got := (*ExtractedInfo)(nil).ThumbnailURL(); got != "" {
		t.Fatalf("nil ThumbnailURL = %q", got)
	}

	preferred := &ExtractedInfo{Thumbnail: new("https://example.com/direct.jpg")}
	if got := preferred.ThumbnailURL(); got != "https://example.com/direct.jpg" {
		t.Fatalf("ThumbnailURL = %q", got)
	}

	info := &ExtractedInfo{
		Thumbnails: []*ExtractedThumbnail{
			nil,
			{URL: "https://example.com/wide.jpg", Width: new(1280)},
			{URL: "https://example.com/pref.jpg", Preference: new(10), Width: new(64)},
			{URL: "https://example.com/pref-wide.jpg", Preference: new(10), Width: new(640)},
			{URL: ""},
		},
	}
	best := info.BestThumbnail()
	if best == nil || best.URL != "https://example.com/pref-wide.jpg" {
		t.Fatalf("best = %+v, want pref then width", best)
	}
}
