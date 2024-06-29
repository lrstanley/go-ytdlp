package template

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Test Variables
var (
	validData       = []byte(`Field1="StringValue"|Field2=123|Field3=45.67`)
	missingData     = []byte(`Field1="StringValue"|Field2=123`)
	malformedData   = []byte(`Field1="StringValue"|Field=123=|Field3=45.67`)
	invalidDataType = []byte(`Field1="StringValue"|Field2=notanumber|Field3=45.67`)
)

type SimpleStruct struct {
	Field1 string
	Field2 int
	Field3 float64
}

type CustomStruct struct {
	Title      string
	TotalBytes int64
	Percent    float64
}

func (c *CustomStruct) UnmarshalTemplate(data []byte) error {
	var raw struct {
		Title      string
		TotalBytes int64
		Percent    string
	}

	if err := Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	c.Title = raw.Title
	c.TotalBytes = raw.TotalBytes

	percent := strings.TrimSpace(strings.TrimSuffix(raw.Percent, "%"))
	percentFloat, err := strconv.ParseFloat(percent, 64)
	if err != nil {
		return fmt.Errorf("failed to parse percent: %w", err)
	}
	c.Percent = percentFloat

	return nil
}

type YtdlpOutput struct {
	ID            string
	Title         string
	Tags          string
	PlaylistIndex int
	Duration      float64
	UploadDate    string
}

var ytdlpTestCases = []struct {
	name     string
	input    string
	expected YtdlpOutput
	hasError bool
}{
	{
		name:  "basic case",
		input: `ID="abc123"|Title="Example Video"|Tags="tag1,tag2"|PlaylistIndex=1|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "Example Video",
			Tags:          "tag1,tag2",
			PlaylistIndex: 1,
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
	{
		name:  "delimiter in value",
		input: `ID="abc123"|Title="Example|Video"|Tags="tag1,tag2"|PlaylistIndex=1|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "Example|Video",
			Tags:          "tag1,tag2",
			PlaylistIndex: 1,
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
	{
		name:  "double quote in value",
		input: `ID="abc123"|Title="Example "Video""|Tags="tag1,tag2"|PlaylistIndex=1|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         `Example "Video"`,
			Tags:          "tag1,tag2",
			PlaylistIndex: 1,
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
	},
	{
		name:  "escaped quote in value",
		input: `ID="abc123"|Title="Example \"Video\""|Tags="tag1,tag2"|PlaylistIndex=1|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         `Example "Video"`,
			Tags:          "tag1,tag2",
			PlaylistIndex: 1,
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
	},
	{
		name:  "integer value",
		input: `ID="abc123"|Title="Example Video"|Tags="tag1,tag2"|PlaylistIndex=42|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "Example Video",
			Tags:          "tag1,tag2",
			PlaylistIndex: 42,
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
	{
		name:  "float value",
		input: `ID="abc123"|Title="Example Video"|Tags="tag1,tag2"|PlaylistIndex=1|Duration=1234.56|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "Example Video",
			Tags:          "tag1,tag2",
			PlaylistIndex: 1,
			Duration:      1234.56,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
	{
		name:  "missing optional field",
		input: `ID="abc123"|Title="Example Video"|Tags="tag1,tag2"|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "Example Video",
			Tags:          "tag1,tag2",
			PlaylistIndex: 0, // default value for int
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
	{
		name:  "empty string field",
		input: `ID="abc123"|Title=""|Tags="tag1,tag2"|PlaylistIndex=1|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "",
			Tags:          "tag1,tag2",
			PlaylistIndex: 1,
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
	{
		name:  "invalid integer value",
		input: `ID="abc123"|Title="Example Video"|Tags="tag1,tag2"|PlaylistIndex=invalid|Duration=3600.5|UploadDate="20230101"`,
		expected: YtdlpOutput{
			ID:            "abc123",
			Title:         "Example Video",
			Tags:          "tag1,tag2",
			PlaylistIndex: 0, // should remain default
			Duration:      3600.5,
			UploadDate:    "20230101",
		},
		hasError: false,
	},
}

func TestDecodeYtdlpOutput(t *testing.T) {
	for _, tc := range ytdlpTestCases {
		t.Run(tc.name, func(t *testing.T) {
			var result YtdlpOutput
			err := Unmarshal([]byte(tc.input), &result)
			if tc.hasError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tc.expected) {
					t.Errorf("expected %+v, got %+v", tc.expected, result)
				}
			}
		})
	}
}

type PercentTagStruct struct {
	Title   string
	Percent float64 `ytdlp:"progress._percent_str,formatter=percentToNumber"`
}

func TestDecode(t *testing.T) {
	type testCase struct {
		name        string
		data        []byte
		v           any
		expected    any
		expectError bool
	}

	testCases := []testCase{
		{
			name:        "unmarshal Non-pointer",
			data:        validData,
			v:           SimpleStruct{},
			expectError: true,
		},
		{
			name:        "unmarshal Nil Pointer",
			data:        validData,
			v:           (*SimpleStruct)(nil),
			expectError: true,
		},
		{
			name: "unmarshal Valid Data",
			data: validData,
			v:    &SimpleStruct{},
			expected: &SimpleStruct{
				Field1: "StringValue",
				Field2: 123,
				Field3: 45.67,
			},
			expectError: false,
		},
		{
			name: "unmarshal Missing Data",
			data: missingData,
			v:    &SimpleStruct{},
			expected: &SimpleStruct{
				Field1: "StringValue",
				Field2: 123,
				Field3: 0.0,
			},
			expectError: false,
		},
		{
			name:        "unmarshal Malformed Data",
			data:        malformedData,
			v:           &SimpleStruct{},
			expectError: true,
		},
		// INFO: ATM, we set the default value for the field if the value is invalid (e.g. not a
		// number), we can add a feature to return an error if the value is invalid.
		// {
		// 	name:        "unmarshal Invalid Data Type",
		// 	data:        invalidDataType,
		// 	v:           &SimpleStruct{},
		// 	expectError: true,
		// },
		{
			name:        "unmarshal with CustomUnmarshaller",
			data:        []byte(`Title="ytdlp"|TotalBytes=123000|Percent="  45.67%"`),
			v:           &CustomStruct{},
			expected:    &CustomStruct{Title: "ytdlp", TotalBytes: 123000, Percent: 45.67},
			expectError: false,
		},
		{
			name: "convert string percent to float when tag is present",
			data: []byte(`Title="ytdlp"|TotalBytes=123000|Percent="  45.67%"`),
			v:    &PercentTagStruct{},
			expected: &PercentTagStruct{
				Title:   "ytdlp",
				Percent: 45.67,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Unmarshal(tc.data, tc.v)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(tc.v, tc.expected) {
					t.Errorf("expected %+v, got %+v", tc.expected, tc.v)
				}
			}
		})
	}
}
