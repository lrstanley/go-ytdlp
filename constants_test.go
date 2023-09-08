// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:forbidigo
package ytdlp

import (
	"testing"
	"time"
)

func TestConstant_Validate(t *testing.T) {
	if Channel == "" {
		t.Fatal("Channel is empty")
	}

	if Version == "" {
		t.Fatal("Version is empty")
	}

	_, err := time.Parse("2006.01.02", Version)
	if err != nil {
		t.Fatalf("failed to parse version: %v", err)
	}
}
