package template

// Formatter is responsible for formatting a string value. e.g. cleaning up a string value, converting a string to a number, etc.
type TemplateFormatter struct {
	formatters map[string]func(string) string
}

// NewTemplateFormatter returns a new TemplateFormatter with the default formatters.
// The default formatters are:
// - percentToNumber: converts a string with a percent to a number.
func NewTemplateFormatter() *TemplateFormatter {
	formatters := make(map[string]func(string) string)
	formatters["percentToNumber"] = func(s string) string {
		return CleanPercent(s)
	}

	return &TemplateFormatter{
		formatters: formatters,
	}
}

// AddFormatter adds a new formatter with the given name and function.
// TODO: we probably want to add a check to see if the formatter already exists.
func (f *TemplateFormatter) AddFormatter(name string, fn func(string) string) {
	f.formatters[name] = fn
}

// Format applies the formatter with the given name to the value.
func (f *TemplateFormatter) Format(name, value string) string {
	if fn, ok := f.formatters[name]; ok {
		return fn(value)
	}
	return value
}
