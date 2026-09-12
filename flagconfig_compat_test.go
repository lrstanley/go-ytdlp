// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json/v2"
	"reflect"
	"strings"
	"testing"
)

func TestFlagConfigUnmarshalJSONWithWarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		assert     func(t *testing.T, config *FlagConfig)
		wantPaths  []string
		wantReason string
	}{
		{
			name:  "unknown members",
			input: `{"unknown":true,"general":{"no_update":true,"unknown_flag":false}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				if config.General.NoUpdate == nil {
					t.Fatal("general.no_update is nil")
				}
				if !*config.General.NoUpdate {
					t.Error("general.no_update = false, want true")
				}
			},
			wantPaths: []string{"general.unknown_flag", "unknown"},
		},
		{
			name:  "valid siblings survive incompatible values",
			input: `{"download":{"concurrent_fragments":"not-a-number","retries":3},"filesystem":{"output":42}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				if config.Download.Retries == nil {
					t.Fatal("download.retries is nil")
				}
				if got := *config.Download.Retries; got != "3" {
					t.Errorf("download.retries = %q, want 3", got)
				}
				if config.Filesystem.Output == nil {
					t.Fatal("filesystem.output is nil")
				}
				if got := *config.Filesystem.Output; got != "42" {
					t.Errorf("filesystem.output = %q, want 42", got)
				}
			},
			wantPaths:  []string{"download.concurrent_fragments"},
			wantReason: "cannot coerce",
		},
		{
			name:  "conservative scalar coercions",
			input: `{"general":{"no_update":"true"},"download":{"concurrent_fragments":"4"},"filesystem":{"output":123}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				if config.General.NoUpdate == nil {
					t.Fatal("general.no_update is nil")
				}
				if !*config.General.NoUpdate {
					t.Error("general.no_update = false, want true")
				}
				if config.Download.ConcurrentFragments == nil {
					t.Fatal("download.concurrent_fragments is nil")
				}
				if got := *config.Download.ConcurrentFragments; got != 4 {
					t.Errorf("download.concurrent_fragments = %d, want 4", got)
				}
				if config.Filesystem.Output == nil {
					t.Fatal("filesystem.output is nil")
				}
				if got := *config.Filesystem.Output; got != "123" {
					t.Errorf("filesystem.output = %q, want 123", got)
				}
			},
		},
		{
			name:  "scalar to singleton slices",
			input: `{"general":{"config_locations":"/tmp/config"},"verbosity_simulation":{"print":42}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				if !reflect.DeepEqual(config.General.ConfigLocations, []string{"/tmp/config"}) {
					t.Errorf("general.config_locations = %v, want [/tmp/config]", config.General.ConfigLocations)
				}
				if !reflect.DeepEqual(config.VerbositySimulation.Print, []string{"42"}) {
					t.Errorf("verbosity_simulation.print = %v, want [42]", config.VerbositySimulation.Print)
				}
			},
		},
		{
			name:  "nested multi argument flag",
			input: `{"verbosity_simulation":{"print_to_file":[{"template":"%(title)s","file":123,"unknown":true}]}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				if len(config.VerbositySimulation.PrintToFile) != 1 {
					t.Fatalf("verbosity_simulation.print_to_file has length %d, want 1", len(config.VerbositySimulation.PrintToFile))
				}
				if got := config.VerbositySimulation.PrintToFile[0].Template; got != "%(title)s" {
					t.Errorf("print_to_file template = %q, want %%(title)s", got)
				}
				if got := config.VerbositySimulation.PrintToFile[0].File; got != "123" {
					t.Errorf("print_to_file file = %q, want 123", got)
				}
			},
			wantPaths:  []string{"verbosity_simulation.print_to_file[0].unknown"},
			wantReason: "unknown member",
		},
		{
			name:  "multiple warnings",
			input: `{"general":{"no_update":"not-a-bool","unknown_one":true,"unknown_two":false}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				if config.General.NoUpdate != nil {
					t.Errorf("general.no_update = %v, want nil", *config.General.NoUpdate)
				}
			},
			wantPaths: []string{
				"general.no_update",
				"general.unknown_one",
				"general.unknown_two",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var config FlagConfig
			warnings, err := config.UnmarshalJSONWithWarnings([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, &config)

			var paths []string
			for _, warning := range warnings {
				paths = append(paths, warning.JSONPath)
				if warning.JSONPath == "download.concurrent_fragments" {
					if warning.Flag != "--concurrent-fragments" {
						t.Errorf("warning flag = %q, want --concurrent-fragments", warning.Flag)
					}
					if warning.ID != "concurrent_fragment_downloads" {
						t.Errorf("warning ID = %q, want concurrent_fragment_downloads", warning.ID)
					}
				}
			}
			if !reflect.DeepEqual(paths, tt.wantPaths) {
				t.Errorf("warning paths = %v, want %v", paths, tt.wantPaths)
			}
			if tt.wantReason != "" {
				if len(warnings) == 0 {
					t.Fatal("expected at least one warning")
				}
				if !strings.Contains(warnings[0].Reason, tt.wantReason) {
					t.Errorf("warning reason = %q, want it to contain %q", warnings[0].Reason, tt.wantReason)
				}
			}
		})
	}
}

func TestFlagConfigUnmarshalJSONWithWarningsCustomCoercion(t *testing.T) {
	t.Parallel()

	var config FlagConfig
	warnings, err := config.UnmarshalJSONWithWarnings(
		[]byte(`{"filesystem":{"output":{"template":"custom"}}}`),
		WithFlagCoercion(func(path string, _ []byte, target reflect.Type) (any, bool, error) {
			if path == "filesystem.output" && target == reflect.TypeOf((*string)(nil)) {
				return "custom-output", true, nil
			}
			return nil, false, nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	if config.Filesystem.Output == nil {
		t.Fatal("filesystem.output is nil")
	}
	if got := *config.Filesystem.Output; got != "custom-output" {
		t.Errorf("filesystem.output = %q, want custom-output", got)
	}
}

func TestFlagConfigUnmarshalJSONWithWarningsInvalidSlicePreservesValue(t *testing.T) {
	t.Parallel()

	config := FlagConfig{
		General: FlagsGeneral{
			ConfigLocations: []string{"existing"},
		},
	}

	warnings, err := config.UnmarshalJSONWithWarnings(
		[]byte(`{"general":{"config_locations":["replacement",{"invalid":true}]}}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings has length %d, want 1", len(warnings))
	}
	if got := warnings[0].JSONPath; got != "general.config_locations[1]" {
		t.Errorf("warning JSON path = %q, want general.config_locations[1]", got)
	}
	if !reflect.DeepEqual(config.General.ConfigLocations, []string{"existing"}) {
		t.Errorf("general.config_locations = %v, want [existing]", config.General.ConfigLocations)
	}
}

func TestFlagConfigUnmarshalJSONWithWarningsMalformedRoot(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"{", "[]", "null"} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			var config FlagConfig
			warnings, err := config.UnmarshalJSONWithWarnings([]byte(input))
			if err == nil {
				t.Fatal("expected an error")
			}
			if len(warnings) != 0 {
				t.Errorf("warnings = %v, want none", warnings)
			}
		})
	}
}

func TestFlagConfigStrictJSONAndClone(t *testing.T) {
	t.Parallel()

	var config FlagConfig
	err := json.Unmarshal([]byte(`{"download":{"concurrent_fragments":"4"}}`), &config)
	if err == nil {
		t.Fatal("expected an error")
	}

	value := true
	original := &FlagConfig{General: FlagsGeneral{NoUpdate: &value}}
	clone := original.Clone()
	if clone.General.NoUpdate == nil {
		t.Fatal("clone general.no_update is nil")
	}
	*clone.General.NoUpdate = false
	if !*original.General.NoUpdate {
		t.Error("original general.no_update = false, want true")
	}
}
