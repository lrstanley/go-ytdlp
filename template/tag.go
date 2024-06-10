package template

import "strings"

type TagParser struct {
	tag   string
	value string

	// formatters is a list of formatters in the tag.
	// e.g. formatter=formatter1,formatter2
	// The formatters are separated by commas.
	// we chose a map to avoid duplicates and for faster lookups.
	formatters map[string]struct{}
}

func NewTagParser(tag string) *TagParser {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return &TagParser{}
	}

	tagSegments := strings.Split(tag, ",")
	value := strings.TrimSpace(tagSegments[0])

	formatters := make(map[string]struct{})
	for _, segment := range tagSegments[1:] {
		segment = strings.TrimSpace(segment)
		if strings.HasPrefix(segment, "formatter=") {
			formatter := strings.TrimPrefix(segment, "formatter=")
			if _, ok := formatters[formatter]; !ok {
				formatters[formatter] = struct{}{}
			}
		}
	}

	return &TagParser{tag: tag, value: value, formatters: formatters}
}

// Tag returns the whole tag.
func (tp *TagParser) Tag() string {
	return tp.tag
}

// IsEmpty returns true if the tag is empty.
func (tp *TagParser) IsEmpty() bool {
	return tp.tag == ""
}

// Value returns the first segment of the tag.
func (tp *TagParser) Value() string {
	return tp.value
}

// Formatters returns the formatters in the tag.
func (tp *TagParser) Formatters() []string {
	var formatters []string
	for k := range tp.formatters {
		formatters = append(formatters, k)
	}
	return formatters
}
