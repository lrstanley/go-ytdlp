// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"testing"
)

var progressTestCases = []struct {
	name  string
	input string
	want  ProgressUpdate
}{
	{
		name:  "downloading-exact-size",
		input: "downloading###4745623###1024###NA###NA###sample-1###NA###NA###NA###https://cdn.liam.sh/github/go-ytdlp/sample-1.mp4###generic - sample-1.mp4",
		want: ProgressUpdate{
			Status:          ProgressStatusDownloading,
			TotalBytes:      4745623,
			DownloadedBytes: 1024,
			ID:              "sample-1",
			URL:             "https://cdn.liam.sh/github/go-ytdlp/sample-1.mp4",
			Filename:        "generic - sample-1.mp4",
		},
	},
	{
		name:  "finished-exact-size",
		input: "finished###4745623###4745623###NA###NA###sample-2###NA###NA###NA###https://cdn.liam.sh/github/go-ytdlp/sample-2.mp4###generic - sample-2.mp4",
		want: ProgressUpdate{
			Status:          ProgressStatusFinished,
			TotalBytes:      4745623,
			DownloadedBytes: 4745623,
			ID:              "sample-2",
			URL:             "https://cdn.liam.sh/github/go-ytdlp/sample-2.mp4",
			Filename:        "generic - sample-2.mp4",
		},
	},
}

func TestProgressHandlerParse(t *testing.T) {
	for _, data := range progressTestCases {
		t.Run(data.name, func(t *testing.T) {
			invoked := false
			h := newProgressHandler(func(update ProgressUpdate) {
				invoked = true

				if update.Status != data.want.Status {
					t.Errorf("expected status %q, got %q", data.want.Status, update.Status)
				}

				if update.TotalBytes != data.want.TotalBytes {
					t.Errorf("expected total bytes %d, got %d", data.want.TotalBytes, update.TotalBytes)
				}

				if update.DownloadedBytes != data.want.DownloadedBytes {
					t.Errorf("expected downloaded bytes %d, got %d", data.want.DownloadedBytes, update.DownloadedBytes)
				}

				if update.ID != data.want.ID {
					t.Errorf("expected id %q, got %q", data.want.ID, update.ID)
				}

				if update.URL != data.want.URL {
					t.Errorf("expected url %q, got %q", data.want.URL, update.URL)
				}

				if update.Filename != data.want.Filename {
					t.Errorf("expected filename %q, got %q", data.want.Filename, update.Filename)
				}
			})

			h.parse(data.input)

			if !invoked {
				t.Error("expected progress callback to be invoked")
			}
		})
	}
}

func FuzzProgressHandlerParse(f *testing.F) {
	for i := range progressTestCases {
		f.Add(progressTestCases[i].input)
	}

	h := newProgressHandler(func(_ ProgressUpdate) {})

	f.Fuzz(func(_ *testing.T, data string) {
		h.parse(data)
	})
}
