// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in the
// LICENSE file.

package ytdlp

import (
	"errors"
	"testing"
)

func TestFlagConfigValidateFlattensGroupErrors(t *testing.T) {
	t.Parallel()

	cfg := &FlagConfig{}
	cfg.General.IgnoreErrors = new(true)
	cfg.General.AbortOnError = new(true)
	cfg.Download.SkipUnavailableFragments = new(true)
	cfg.Download.AbortOnUnavailableFragments = new(true)

	err := cfg.Validate()
	flags := JSONParsingFlagErrors(err)
	if len(flags) == 0 {
		t.Fatalf("Validate() = %v, want JSON parsing flag errors", err)
	}

	ids := map[string]int{}
	for _, e := range flags {
		ids[e.ID]++
	}
	if ids["ignoreerrors"] == 0 {
		t.Fatalf("missing general ignoreerrors conflicts: %+v", flags)
	}
	if ids["skip_unavailable_fragments"] == 0 {
		t.Fatalf("missing download skip_unavailable_fragments conflicts: %+v", flags)
	}

	var flagErr *ErrJSONParsingFlag
	if !errors.As(err, &flagErr) {
		t.Fatal("errors.As should find ErrJSONParsingFlag via Unwrap")
	}
	if !errors.Is(err, flags[0]) {
		t.Fatal("errors.Is should match an inner flag error")
	}
}

func TestFlagConfigRepairUnsetsAllConflictIDs(t *testing.T) {
	t.Parallel()

	cfg := &FlagConfig{}
	cfg.General.IgnoreErrors = new(true)
	cfg.General.AbortOnError = new(true)
	cfg.Download.SkipUnavailableFragments = new(true)
	cfg.Download.AbortOnUnavailableFragments = new(true)
	cfg.Filesystem.Output = new("out.%(ext)s")

	warnings := cfg.Repair()
	if len(warnings) == 0 {
		t.Fatal("expected repair warnings")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("repaired config still invalid: %v", err)
	}
	if cfg.General.IgnoreErrors != nil || cfg.General.AbortOnError != nil {
		t.Fatal("ignoreerrors siblings were not both unset")
	}
	if cfg.Download.SkipUnavailableFragments != nil || cfg.Download.AbortOnUnavailableFragments != nil {
		t.Fatal("skip_unavailable_fragments siblings were not both unset")
	}
	if cfg.Filesystem.Output == nil || *cfg.Filesystem.Output != "out.%(ext)s" {
		t.Fatal("unrelated field was cleared")
	}
}

func TestFlagConfigUnmarshalRepairEmpty(t *testing.T) {
	t.Parallel()

	for _, data := range [][]byte{nil, {}, []byte(""), []byte("  \n")} {
		var cfg FlagConfig
		if _, err := cfg.UnmarshalRepair(data); err != nil {
			t.Fatalf("UnmarshalRepair(%q) err = %v", data, err)
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("empty config invalid: %v", err)
		}
	}
}

func TestFlagConfigUnmarshalRepairHardFailLeavesReceiver(t *testing.T) {
	t.Parallel()

	cfg := FlagConfig{Filesystem: FlagsFilesystem{Output: new("keep")}}
	_, err := cfg.UnmarshalRepair([]byte(`[1,2,3]`))
	if err == nil {
		t.Fatal("expected hard fail")
	}
	if cfg.Filesystem.Output == nil || *cfg.Filesystem.Output != "keep" {
		t.Fatal("receiver was modified on hard fail")
	}
}

func TestFlagConfigOverlay(t *testing.T) {
	t.Parallel()

	base := &FlagConfig{}
	base.Filesystem.Output = new("base.%(ext)s")
	base.Filesystem.Continue = new(true)
	base.General.ConfigLocations = []string{"/base.conf"}
	base.Download.Retries = new("3")

	over := &FlagConfig{}
	over.Filesystem.Output = new("over.%(ext)s")
	over.Filesystem.Continue = new(false)
	over.General.ConfigLocations = []string{"/over.conf"}

	got := base.Overlay(over)
	if got.Filesystem.Output == nil || *got.Filesystem.Output != "over.%(ext)s" {
		t.Fatalf("output = %v, want over", got.Filesystem.Output)
	}
	if got.Filesystem.Continue == nil || *got.Filesystem.Continue {
		t.Fatal("continue should be overlay false")
	}
	if got.Download.Retries == nil || *got.Download.Retries != "3" {
		t.Fatal("nil overlay field cleared base retries")
	}
	if len(got.General.ConfigLocations) != 1 || got.General.ConfigLocations[0] != "/over.conf" {
		t.Fatalf("config_locations = %v, want overlay", got.General.ConfigLocations)
	}
	if *base.Filesystem.Output != "base.%(ext)s" {
		t.Fatal("overlay mutated base")
	}

	emptySlice := &FlagConfig{}
	emptySlice.General.ConfigLocations = []string{}
	kept := base.Overlay(emptySlice)
	if len(kept.General.ConfigLocations) != 1 {
		t.Fatal("empty overlay slice wiped base slice")
	}
}

func TestOverlayFlagConfig(t *testing.T) {
	t.Parallel()

	cfg, warnings, err := OverlayFlagConfig(
		[]byte(`{"filesystem":{"output":"base.%(ext)s","continue":true},"download":{"skip_unavailable_fragments":true,"abort_on_unavailable_fragments":true}}`),
		[]byte(`{"filesystem":{"output":"over.%(ext)s"}}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Filesystem.Output == nil || *cfg.Filesystem.Output != "over.%(ext)s" {
		t.Fatalf("output = %v", cfg.Filesystem.Output)
	}
	if cfg.Filesystem.Continue == nil || !*cfg.Filesystem.Continue {
		t.Fatal("base continue was dropped")
	}
	if err = cfg.Validate(); err != nil {
		t.Fatalf("merged config invalid: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected conflict warnings from base")
	}
}
