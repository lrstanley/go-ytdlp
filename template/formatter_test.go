package template

import (
	"testing"
)

func TestNewTemplateFormatter(t *testing.T) {
	formatter := NewTemplateFormatter()
	if _, exists := formatter.formatters["percentToNumber"]; !exists {
		t.Errorf("expected default formatter 'percentToNumber' to exist")
	}
}

func TestTemplateFormatter_AddFormatter(t *testing.T) {
	formatter := NewTemplateFormatter()
	customFormatterName := "customFormat"
	customFormatter := func(s string) string {
		return "custom formatted: " + s
	}

	formatter.AddFormatter(customFormatterName, customFormatter)
	if _, exists := formatter.formatters[customFormatterName]; !exists {
		t.Errorf("expected custom formatter '%s' to be added", customFormatterName)
	}
}

func TestTemplateFormatter_Format(t *testing.T) {
	formatter := NewTemplateFormatter()
	customFormatterName := "customFormat"
	customFormatter := func(s string) string {
		return "custom formatted: " + s
	}
	formatter.AddFormatter(customFormatterName, customFormatter)

	tests := []struct {
		name     string
		format   string
		value    string
		expected string
	}{
		{"Default Percent Formatter", "percentToNumber", " 50% ", "50"},
		{"Custom Formatter", customFormatterName, "test", "custom formatted: test"},
		{"Nonexistent Formatter", "nonexistentFormat", "test", "test"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := formatter.Format(test.format, test.value)
			if result != test.expected {
				t.Errorf("Format(%q, %q) = %q; want %q", test.format, test.value, result, test.expected)
			}
		})
	}
}
