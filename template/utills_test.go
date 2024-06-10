package template

import "testing"

func TestCleanPercent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  50%  ", "50"},
		{"50%", "50"},
		{"  75%  ", "75"},
		{"%", ""},
		{"   ", ""},
		{"", ""},
		{"100", "100"},
		{"  20  ", "20"},
		{"50 %%", "50"},
		{"%50%", "50"},

		{"  85 %", "85"},
		{" % 90 % ", "90"},
		{"100 % ", "100"},
	}

	for _, test := range tests {
		result := CleanPercent(test.input)
		if result != test.expected {
			t.Errorf("CleanPercent(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}
