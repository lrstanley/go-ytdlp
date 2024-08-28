// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"testing"
)

var progressFuzzExamples = []string{
	"downloading,4745623,NA,1024,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,3072,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,31744,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,64512,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,130048,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,261120,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,523264,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,1023037,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,1522880,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,2022722,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,2522905,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,3022712,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,3522839,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,4022890,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,4522676,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"downloading,4745623,NA,4745623,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
	"finished,4745623,NA,4745623,NA,NA,sample-1,NA,NA,NA,generic - sample-1.mp4",
}

func FuzzProgressHandlerParse(f *testing.F) {
	for _, data := range progressFuzzExamples {
		f.Add(data)
	}

	h := newProgressHandler(func(_ ProgressUpdate) {})

	f.Fuzz(func(t *testing.T, data string) {
		h.parse(data)
	})
}
