// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:forbidigo
package ytdlp

import (
	"testing"
	"time"
)

func TestConstant_ValidateMain(t *testing.T) {
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

func TestConstant_ValidateExtractors(t *testing.T) {
	if len(SupportedExtractors) == 0 {
		t.Fatal("SupportedExtractors is empty")
	}

	withDescriptions := 0
	withAgeLimit := 0
	for _, e := range SupportedExtractors {
		if e.Name == "" {
			t.Fatal("extractor has no name")
		}

		if e.Description != "" {
			withDescriptions++
		}

		if e.AgeLimit != 0 {
			withAgeLimit++
		}
	}

	if withDescriptions == 0 {
		t.Fatal("no extractors have descriptions")
	}

	if withAgeLimit == 0 {
		t.Fatal("no extractors have age limits")
	}
}
