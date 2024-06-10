package template

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrInvalidUnmarshalError = errors.New("ytdlp: Unmarshal(non-pointer)")
	ErrUnmarshalTypeError    = errors.New("ytdlp: cannot unmarshal")
	ErrMalformedDataError    = errors.New("ytdlp: malformed data")
)

// Unmarshaller is the interface implemented by types that can unmarshal a delimiter-separated value description of themselves.
type Unmarshaller interface {
	UnmarshalTemplate([]byte) error
}

// decodeState holds the state while decoding a delimiter-separated format.
type decodeState struct {
	data      []byte
	str       string
	index     int
	decoder   *Decoder
	formatter *TemplateFormatter
}

// Decoder reads and decodes a YTDLP element from an input stream.
type Decoder struct {
	r         *bytes.Buffer
	formatter *TemplateFormatter
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r *bytes.Buffer, formatter *TemplateFormatter) *Decoder {
	return &Decoder{r: r, formatter: formatter}
}

// Unmarshal parses the YTDLP-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	d := NewDecoder(bytes.NewBuffer(data), NewTemplateFormatter())

	if !d.isValid() {
		return fmt.Errorf("ytdlp: invalid data")
	}

	return d.Decode(v)
}

func (d *Decoder) isValid() bool {
	return d.r.Len() > 0
}

// Decode reads the YTDLP-encoded value from its input and stores it in the value pointed to by v.
func (d *Decoder) Decode(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return ErrInvalidUnmarshalError
	}

	// If the formatter is nil, use the default formatter
	if d.formatter == nil {
		d.formatter = NewTemplateFormatter()
	}

	dstate := &decodeState{
		data:      d.r.Bytes(),
		str:       d.r.String(),
		index:     0,
		decoder:   d,
		formatter: d.formatter,
	}
	return dstate.unmarshal(v)
}

func (d *decodeState) unmarshal(v any) error {
	val := reflect.ValueOf(v).Elem()
	typ := val.Type()

	// Check if the type implements the Unmarshaller interface
	if um, ok := v.(Unmarshaller); ok {
		return um.UnmarshalTemplate(d.data)
	}

	fieldMap := make(map[string]string)

	// Split the input data into key-value pairs based on the delimiter
	var currentKey, currentValue strings.Builder
	isInQuotes := false

	for i := 0; i < len(d.str); i++ {
		char := d.str[i]

		switch char {
		case '"':
			if i > 0 && d.str[i-1] == '\\' {
				currentValue.WriteByte(char)
			} else {
				isInQuotes = !isInQuotes
				currentValue.WriteByte(char)
			}
		case '=':
			if !isInQuotes {
				if currentKey.Len() > 0 {
					return fmt.Errorf("%w: unexpected '=' character", ErrMalformedDataError)
				}
				currentKey.WriteString(strings.TrimSpace(currentValue.String()))
				currentValue.Reset()
			} else {
				currentValue.WriteByte(char)
			}
		case '|':
			if !isInQuotes {
				if currentKey.Len() == 0 || currentValue.Len() == 0 {
					return fmt.Errorf("%w: incomplete key-value pair", ErrMalformedDataError)
				}
				fieldMap[currentKey.String()] = strings.TrimSpace(currentValue.String())
				currentKey.Reset()
				currentValue.Reset()
			} else {
				currentValue.WriteByte(char)
			}
		default:
			currentValue.WriteByte(char)
		}
	}

	if isInQuotes {
		return fmt.Errorf("%w: unbalanced quotes", ErrMalformedDataError)
	}

	if currentKey.Len() > 0 && currentValue.Len() > 0 {
		fieldMap[currentKey.String()] = currentValue.String()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)
		tag := NewTagParser(structField.Tag.Get("ytdlp"))

		if fieldValue, ok := fieldMap[structField.Name]; ok {
			if len(fieldValue) >= 2 && fieldValue[0] == '"' && fieldValue[len(fieldValue)-1] == '"' {
				fieldValue = fieldValue[1 : len(fieldValue)-1] // Trim only one leading and trailing quote
			}
			fieldValue = strings.ReplaceAll(fieldValue, "\\\"", "\"") // Unescape quotes

			formatters := tag.Formatters()
			for _, formatterName := range formatters {
				fieldValue = d.formatter.Format(formatterName, fieldValue)
			}

			switch field.Kind() {
			case reflect.String:
				field.SetString(fieldValue)
			case reflect.Bool:
				if boolValue, err := strconv.ParseBool(fieldValue); err == nil {
					field.SetBool(boolValue)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if intValue, err := strconv.Atoi(fieldValue); err == nil {
					field.SetInt(int64(intValue))
				}
			case reflect.Float32, reflect.Float64:
				if floatValue, err := strconv.ParseFloat(fieldValue, 64); err == nil {
					field.SetFloat(floatValue)
				}
			default:
				return fmt.Errorf("%w: %s into %s", ErrUnmarshalTypeError, fieldValue, field.Type())
			}
		}
	}
	return nil
}

