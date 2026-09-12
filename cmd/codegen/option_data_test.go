// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"slices"
	"testing"
)

func TestOptionGroupGenerateMatchesAllFlags(t *testing.T) {
	group := OptionGroup{
		Options: []Option{
			{
				ID:        "example",
				LongFlags: []string{"--canonical", "--help"},
			},
		},
	}

	group.Generate(&OptionData{})

	if len(group.Options) != 0 {
		t.Fatalf("expected ignored alias to remove option, got %d options", len(group.Options))
	}
}

func TestOptionGenerateMatchesDeprecatedAliases(t *testing.T) {
	option := Option{
		ID:        "writesubtitles",
		LongFlags: []string{"--write-subs", "--write-srt"},
	}

	option.Generate(&OptionGroup{Parent: &OptionData{}})

	if len(option.DeprecatedAliases) != 1 {
		t.Fatalf("expected one deprecated alias, got %d", len(option.DeprecatedAliases))
	}
	if option.DeprecatedAliases[0].Flag != "--write-srt" {
		t.Fatalf("unexpected deprecated alias: %s", option.DeprecatedAliases[0].Flag)
	}
}

func TestOptionDataStalePoliciesChecksAllFlags(t *testing.T) {
	data := OptionData{
		OptionGroups: []OptionGroup{
			{
				Options: []Option{
					{
						ID:        "format_sort",
						LongFlags: []string{"--format-sort", "--format-sort-reset"},
					},
					{
						ID:        "writesubtitles",
						LongFlags: []string{"--write-subs", "--write-srt"},
					},
					{
						ID: "version",
					},
				},
			},
		},
	}

	stale := data.StalePolicies()
	for _, policy := range []string{
		"ignored flag: --format-sort-reset",
		"deprecated flag: --write-srt",
		"executable option: version",
	} {
		if slices.Contains(stale, policy) {
			t.Errorf("unexpected stale policy: %s", policy)
		}
	}
	if err := data.ValidatePolicies(); err == nil {
		t.Fatal("expected stale policies to fail validation")
	}
}
