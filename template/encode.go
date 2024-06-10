package template

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

const templateDelimiter = "|"

// Encoder writes YTDLP elements to an output stream.
type Encoder struct {
	w *bytes.Buffer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w *bytes.Buffer) *Encoder {
	return &Encoder{w: w}
}

// MarshalYTDLP returns the YTDLP encoding of v.
func MarshalTemplate(v any) ([]byte, error) {
	var buf bytes.Buffer
	e := NewEncoder(&buf)
	err := e.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode writes the YTDLP encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ytdlp: Marshal(non-struct %s)", val.Type())
	}

	var sb strings.Builder
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)
		tag := NewTagParser(structField.Tag.Get("ytdlp"))

		// Skip fields with empty tags and zero values.
		if tag.IsEmpty() && field.IsZero() {
			continue
		}

		var fieldStr strings.Builder
		fieldStr.WriteString(fmt.Sprintf("%s=", structField.Name))

		// https://github.com/yt-dlp/yt-dlp?tab=readme-ov-file#output-template
		switch field.Kind() {
		case reflect.String:
			if tag.IsEmpty() {
				fieldStr.WriteString(fmt.Sprintf("%q", field.String()))
			} else {
				fieldStr.WriteString(fmt.Sprintf("%%(%s|\"\")j", tag.Value()))
			}
		case reflect.Bool:
			if tag.IsEmpty() {
				fieldStr.WriteString(fmt.Sprintf("%t", field.Bool()))
			} else {
				fieldStr.WriteString(fmt.Sprintf("%%(%s|false)j", tag.Value()))
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
			if tag.IsEmpty() {
				fieldStr.WriteString(fmt.Sprintf("%d", field.Int()))
			} else {
				fieldStr.WriteString(fmt.Sprintf("%%(%s|0)j", tag.Value()))
			}
		default:
			continue
		}

		if sb.Len() > 0 {
			sb.WriteString(templateDelimiter)
		}
		sb.WriteString(fieldStr.String())
	}

	e.w.WriteString(sb.String())
	return nil
}
