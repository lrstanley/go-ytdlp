// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package optiondata

import (
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaOptionDataDefinitions(t *testing.T) {
	t.Parallel()

	var document map[string]any
	require.NoError(t, json.Unmarshal(JSONSchema, &document))

	assert.Equal(t, "#/$defs/FlagConfig", document["$ref"])

	definitions, ok := document["$defs"].(map[string]any)
	require.True(t, ok)
	for _, name := range []string{"Option", "OptionGroup", "OptionURL"} {
		require.Contains(t, definitions, name)
	}

	flagConfig := schemaDefinition(t, definitions, "FlagConfig")
	assert.Equal(t, "object", flagConfig["type"])
	assert.Equal(t, false, flagConfig["additionalProperties"])
	flagConfigProperties := schemaProperties(t, flagConfig)
	assert.Equal(t, map[string]string{
		"general":              "#/$defs/FlagsGeneral",
		"network":              "#/$defs/FlagsNetwork",
		"geo_restriction":      "#/$defs/FlagsGeoRestriction",
		"video_selection":      "#/$defs/FlagsVideoSelection",
		"download":             "#/$defs/FlagsDownload",
		"filesystem":           "#/$defs/FlagsFilesystem",
		"thumbnail":            "#/$defs/FlagsThumbnail",
		"internet_shortcut":    "#/$defs/FlagsInternetShortcut",
		"verbosity_simulation": "#/$defs/FlagsVerbositySimulation",
		"workarounds":          "#/$defs/FlagsWorkarounds",
		"video_format":         "#/$defs/FlagsVideoFormat",
		"subtitle":             "#/$defs/FlagsSubtitle",
		"authentication":       "#/$defs/FlagsAuthentication",
		"post_processing":      "#/$defs/FlagsPostProcessing",
		"sponsor_block":        "#/$defs/FlagsSponsorBlock",
		"extractor":            "#/$defs/FlagsExtractor",
	}, schemaPropertyRefs(t, flagConfigProperties))

	option := schemaDefinition(t, definitions, "Option")
	assert.Equal(t, "object", option["type"])
	assert.Equal(t, false, option["additionalProperties"])
	assert.ElementsMatch(t, []string{
		"name",
		"name_camel_case",
		"name_pascal_case",
		"name_snake_case",
		"default_flag",
		"executable",
		"choices",
		"hidden",
		"type",
		"long_flags",
		"short_flags",
		"nargs",
		"no_override",
	}, schemaRequired(t, option))
	optionProperties := schemaProperties(t, option)
	assert.Equal(t, "string", schemaProperty(t, optionProperties, "name")["type"])
	assert.Equal(t, "boolean", schemaProperty(t, optionProperties, "executable")["type"])
	assert.Equal(t, "integer", schemaProperty(t, optionProperties, "nargs")["type"])
	assert.Equal(t, "array", schemaProperty(t, optionProperties, "urls")["type"])
	assert.Equal(t, "#/$defs/OptionURL", schemaItems(t, schemaProperty(t, optionProperties, "urls"))["$ref"])

	optionGroup := schemaDefinition(t, definitions, "OptionGroup")
	assert.Equal(t, "object", optionGroup["type"])
	assert.Equal(t, false, optionGroup["additionalProperties"])
	assert.ElementsMatch(t, []string{"name", "options"}, schemaRequired(t, optionGroup))
	groupProperties := schemaProperties(t, optionGroup)
	assert.Equal(t, "string", schemaProperty(t, groupProperties, "name")["type"])
	assert.Equal(t, "array", schemaProperty(t, groupProperties, "options")["type"])
	assert.Equal(t, "#/$defs/Option", schemaItems(t, schemaProperty(t, groupProperties, "options"))["$ref"])

	optionURL := schemaDefinition(t, definitions, "OptionURL")
	assert.Equal(t, "object", optionURL["type"])
	assert.Equal(t, false, optionURL["additionalProperties"])
	assert.ElementsMatch(t, []string{"name", "url"}, schemaRequired(t, optionURL))
	urlProperties := schemaProperties(t, optionURL)
	assert.Equal(t, "string", schemaProperty(t, urlProperties, "name")["type"])
	assert.Equal(t, "string", schemaProperty(t, urlProperties, "url")["type"])

	var refs []string
	collectSchemaRefs(document, &refs)
	for _, ref := range refs {
		target, found := strings.CutPrefix(ref, "#/$defs/")
		require.Truef(t, found, "unexpected internal reference %q", ref)
		require.Contains(t, definitions, target, "unresolved reference %q", ref)
	}
}

func schemaDefinition(t *testing.T, definitions map[string]any, name string) map[string]any {
	t.Helper()

	definition, ok := definitions[name].(map[string]any)
	require.Truef(t, ok, "definition %q is not an object", name)
	return definition
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	return properties
}

func schemaProperty(t *testing.T, properties map[string]any, name string) map[string]any {
	t.Helper()

	property, ok := properties[name].(map[string]any)
	require.Truef(t, ok, "property %q is not an object", name)
	return property
}

func schemaPropertyRefs(t *testing.T, properties map[string]any) map[string]string {
	t.Helper()

	refs := make(map[string]string, len(properties))
	for name, value := range properties {
		property, ok := value.(map[string]any)
		require.Truef(t, ok, "property %q is not an object", name)
		ref, ok := property["$ref"].(string)
		require.Truef(t, ok, "property %q has no reference", name)
		refs[name] = ref
	}
	return refs
}

func schemaItems(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	items, ok := schema["items"].(map[string]any)
	require.True(t, ok)
	return items
}

func schemaRequired(t *testing.T, schema map[string]any) []string {
	t.Helper()

	required, ok := schema["required"].([]any)
	require.True(t, ok)
	result := make([]string, 0, len(required))
	for _, name := range required {
		value, isString := name.(string)
		require.True(t, isString)
		result = append(result, value)
	}
	return result
}

func collectSchemaRefs(value any, refs *[]string) {
	switch value := value.(type) {
	case map[string]any:
		if ref, ok := value["$ref"].(string); ok {
			*refs = append(*refs, ref)
		}
		for _, child := range value {
			collectSchemaRefs(child, refs)
		}
	case []any:
		for _, child := range value {
			collectSchemaRefs(child, refs)
		}
	}
}
