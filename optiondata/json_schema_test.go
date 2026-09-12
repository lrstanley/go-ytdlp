// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package optiondata

import (
	"encoding/json/v2"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestJSONSchemaOptionDataDefinitions(t *testing.T) {
	t.Parallel()

	var document map[string]any
	if err := json.Unmarshal(JSONSchema, &document); err != nil {
		t.Fatal(err)
	}

	if got := document["$ref"]; got != "#/$defs/FlagConfig" {
		t.Errorf("$ref = %v, want #/$defs/FlagConfig", got)
	}

	definitions, ok := document["$defs"].(map[string]any)
	if !ok {
		t.Fatal("$defs is not an object")
	}
	for _, name := range []string{"Option", "OptionGroup", "OptionURL"} {
		if _, found := definitions[name]; !found {
			t.Errorf("$defs does not contain %q", name)
		}
	}

	flagConfig := schemaDefinition(t, definitions, "FlagConfig")
	if got := flagConfig["type"]; got != "object" {
		t.Errorf("FlagConfig type = %v, want object", got)
	}
	if got := flagConfig["additionalProperties"]; got != false {
		t.Errorf("FlagConfig additionalProperties = %v, want false", got)
	}
	flagConfigProperties := schemaProperties(t, flagConfig)
	if got, want := schemaPropertyRefs(t, flagConfigProperties), map[string]string{
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
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("FlagConfig properties = %v, want %v", got, want)
	}

	option := schemaDefinition(t, definitions, "Option")
	if got := option["type"]; got != "object" {
		t.Errorf("Option type = %v, want object", got)
	}
	if got := option["additionalProperties"]; got != false {
		t.Errorf("Option additionalProperties = %v, want false", got)
	}
	gotRequired := schemaRequired(t, option)
	wantRequired := []string{
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
	}
	slices.Sort(gotRequired)
	slices.Sort(wantRequired)
	if !slices.Equal(gotRequired, wantRequired) {
		t.Errorf("Option required = %v, want %v", gotRequired, wantRequired)
	}
	optionProperties := schemaProperties(t, option)
	if got := schemaProperty(t, optionProperties, "name")["type"]; got != "string" {
		t.Errorf("Option name type = %v, want string", got)
	}
	if got := schemaProperty(t, optionProperties, "executable")["type"]; got != "boolean" {
		t.Errorf("Option executable type = %v, want boolean", got)
	}
	if got := schemaProperty(t, optionProperties, "nargs")["type"]; got != "integer" {
		t.Errorf("Option nargs type = %v, want integer", got)
	}
	if got := schemaProperty(t, optionProperties, "urls")["type"]; got != "array" {
		t.Errorf("Option urls type = %v, want array", got)
	}
	if got := schemaItems(t, schemaProperty(t, optionProperties, "urls"))["$ref"]; got != "#/$defs/OptionURL" {
		t.Errorf("Option urls items ref = %v, want #/$defs/OptionURL", got)
	}

	optionGroup := schemaDefinition(t, definitions, "OptionGroup")
	if got := optionGroup["type"]; got != "object" {
		t.Errorf("OptionGroup type = %v, want object", got)
	}
	if got := optionGroup["additionalProperties"]; got != false {
		t.Errorf("OptionGroup additionalProperties = %v, want false", got)
	}
	gotRequired = schemaRequired(t, optionGroup)
	wantRequired = []string{"name", "options"}
	slices.Sort(gotRequired)
	slices.Sort(wantRequired)
	if !slices.Equal(gotRequired, wantRequired) {
		t.Errorf("OptionGroup required = %v, want %v", gotRequired, wantRequired)
	}
	groupProperties := schemaProperties(t, optionGroup)
	if got := schemaProperty(t, groupProperties, "name")["type"]; got != "string" {
		t.Errorf("OptionGroup name type = %v, want string", got)
	}
	if got := schemaProperty(t, groupProperties, "options")["type"]; got != "array" {
		t.Errorf("OptionGroup options type = %v, want array", got)
	}
	if got := schemaItems(t, schemaProperty(t, groupProperties, "options"))["$ref"]; got != "#/$defs/Option" {
		t.Errorf("OptionGroup options items ref = %v, want #/$defs/Option", got)
	}

	optionURL := schemaDefinition(t, definitions, "OptionURL")
	if got := optionURL["type"]; got != "object" {
		t.Errorf("OptionURL type = %v, want object", got)
	}
	if got := optionURL["additionalProperties"]; got != false {
		t.Errorf("OptionURL additionalProperties = %v, want false", got)
	}
	gotRequired = schemaRequired(t, optionURL)
	wantRequired = []string{"name", "url"}
	slices.Sort(gotRequired)
	slices.Sort(wantRequired)
	if !slices.Equal(gotRequired, wantRequired) {
		t.Errorf("OptionURL required = %v, want %v", gotRequired, wantRequired)
	}
	urlProperties := schemaProperties(t, optionURL)
	if got := schemaProperty(t, urlProperties, "name")["type"]; got != "string" {
		t.Errorf("OptionURL name type = %v, want string", got)
	}
	if got := schemaProperty(t, urlProperties, "url")["type"]; got != "string" {
		t.Errorf("OptionURL url type = %v, want string", got)
	}

	var refs []string
	collectSchemaRefs(document, &refs)
	for _, ref := range refs {
		target, found := strings.CutPrefix(ref, "#/$defs/")
		if !found {
			t.Fatalf("unexpected internal reference %q", ref)
		}
		if _, found := definitions[target]; !found {
			t.Fatalf("unresolved reference %q", ref)
		}
	}
}

func schemaDefinition(t *testing.T, definitions map[string]any, name string) map[string]any {
	t.Helper()

	definition, ok := definitions[name].(map[string]any)
	if !ok {
		t.Fatalf("definition %q is not an object", name)
	}
	return definition
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema properties is not an object")
	}
	return properties
}

func schemaProperty(t *testing.T, properties map[string]any, name string) map[string]any {
	t.Helper()

	property, ok := properties[name].(map[string]any)
	if !ok {
		t.Fatalf("property %q is not an object", name)
	}
	return property
}

func schemaPropertyRefs(t *testing.T, properties map[string]any) map[string]string {
	t.Helper()

	refs := make(map[string]string, len(properties))
	for name, value := range properties {
		property, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("property %q is not an object", name)
		}
		ref, ok := property["$ref"].(string)
		if !ok {
			t.Fatalf("property %q has no reference", name)
		}
		refs[name] = ref
	}
	return refs
}

func schemaItems(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	items, ok := schema["items"].(map[string]any)
	if !ok {
		t.Fatal("schema items is not an object")
	}
	return items
}

func schemaRequired(t *testing.T, schema map[string]any) []string {
	t.Helper()

	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("schema required is not an array")
	}
	result := make([]string, 0, len(required))
	for _, name := range required {
		value, isString := name.(string)
		if !isString {
			t.Fatalf("schema required member %v is not a string", name)
		}
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
