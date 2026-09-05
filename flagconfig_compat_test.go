// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json/v2"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NotNil(t, config.General.NoUpdate)
				assert.True(t, *config.General.NoUpdate)
			},
			wantPaths: []string{"general.unknown_flag", "unknown"},
		},
		{
			name:  "valid siblings survive incompatible values",
			input: `{"download":{"concurrent_fragments":"not-a-number","retries":3},"filesystem":{"output":42}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				require.NotNil(t, config.Download.Retries)
				assert.Equal(t, "3", *config.Download.Retries)
				require.NotNil(t, config.Filesystem.Output)
				assert.Equal(t, "42", *config.Filesystem.Output)
			},
			wantPaths:  []string{"download.concurrent_fragments"},
			wantReason: "cannot coerce",
		},
		{
			name:  "conservative scalar coercions",
			input: `{"general":{"no_update":"true"},"download":{"concurrent_fragments":"4"},"filesystem":{"output":123}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				require.NotNil(t, config.General.NoUpdate)
				assert.True(t, *config.General.NoUpdate)
				require.NotNil(t, config.Download.ConcurrentFragments)
				assert.Equal(t, 4, *config.Download.ConcurrentFragments)
				require.NotNil(t, config.Filesystem.Output)
				assert.Equal(t, "123", *config.Filesystem.Output)
			},
		},
		{
			name:  "scalar to singleton slices",
			input: `{"general":{"config_locations":"/tmp/config"},"verbosity_simulation":{"print":42}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				assert.Equal(t, []string{"/tmp/config"}, config.General.ConfigLocations)
				assert.Equal(t, []string{"42"}, config.VerbositySimulation.Print)
			},
		},
		{
			name:  "nested multi argument flag",
			input: `{"verbosity_simulation":{"print_to_file":[{"template":"%(title)s","file":123,"unknown":true}]}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				require.Len(t, config.VerbositySimulation.PrintToFile, 1)
				assert.Equal(t, "%(title)s", config.VerbositySimulation.PrintToFile[0].Template)
				assert.Equal(t, "123", config.VerbositySimulation.PrintToFile[0].File)
			},
			wantPaths:  []string{"verbosity_simulation.print_to_file[0].unknown"},
			wantReason: "unknown member",
		},
		{
			name:  "multiple warnings",
			input: `{"general":{"no_update":"not-a-bool","unknown_one":true,"unknown_two":false}}`,
			assert: func(t *testing.T, config *FlagConfig) {
				assert.Nil(t, config.General.NoUpdate)
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
			require.NoError(t, err)
			tt.assert(t, &config)

			var paths []string
			for _, warning := range warnings {
				paths = append(paths, warning.JSONPath)
				if warning.JSONPath == "download.concurrent_fragments" {
					assert.Equal(t, "--concurrent-fragments", warning.Flag)
					assert.Equal(t, "concurrent_fragment_downloads", warning.ID)
				}
			}
			assert.Equal(t, tt.wantPaths, paths)
			if tt.wantReason != "" {
				assert.Contains(t, warnings[0].Reason, tt.wantReason)
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
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.NotNil(t, config.Filesystem.Output)
	assert.Equal(t, "custom-output", *config.Filesystem.Output)
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
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Equal(t, "general.config_locations[1]", warnings[0].JSONPath)
	assert.Equal(t, []string{"existing"}, config.General.ConfigLocations)
}

func TestFlagConfigUnmarshalJSONWithWarningsMalformedRoot(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"{", "[]", "null"} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			var config FlagConfig
			warnings, err := config.UnmarshalJSONWithWarnings([]byte(input))
			require.Error(t, err)
			assert.Empty(t, warnings)
		})
	}
}

func TestFlagConfigStrictJSONAndClone(t *testing.T) {
	t.Parallel()

	var config FlagConfig
	err := json.Unmarshal([]byte(`{"download":{"concurrent_fragments":"4"}}`), &config)
	require.Error(t, err)

	value := true
	original := &FlagConfig{General: FlagsGeneral{NoUpdate: &value}}
	clone := original.Clone()
	require.NotNil(t, clone.General.NoUpdate)
	*clone.General.NoUpdate = false
	assert.True(t, *original.General.NoUpdate)
}
