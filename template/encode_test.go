package template

import (
	"bytes"
	"testing"
)

type TestStruct struct {
	Field1 string            `ytdlp:"field1"`
	Field2 int               `ytdlp:"field2"`
	Field3 bool              `ytdlp:"field3"`
	Field4 map[string]string `ytdlp:"field4"` // unsupported type
}

func TestMarshalTemplate(t *testing.T) {
	t.Run("marshal non struct", func(t *testing.T) {
		_, err := MarshalTemplate("string value")
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
	})

	t.Run("marshal struct with unsupported type", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		err := enc.Encode(TestStruct{Field1: "test", Field2: 123, Field3: true, Field4: map[string]string{"key": "value"}})
		if err != nil {
			t.Fatalf("expected error but got none")
		}
		expected := "Field1=%(field1|\"\")j|Field2=%(field2|0)j|Field3=%(field3|false)j"
		if buf.String() != expected {
			t.Errorf("expected %q but got %q", expected, buf.String())
		}
	})

	t.Run("marshal empty struct with Ytdlp tags", func(t *testing.T) {
		type EmptyStruct struct {
			Field1 string `ytdlp:"field1"`
		}

		expected := "Field1=%(field1|\"\")j"
		data, err := MarshalTemplate(EmptyStruct{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != expected {
			t.Errorf("expected %q but got %q", expected, data)
		}
	})

	t.Run("marshal struct", func(t *testing.T) {
		type SimpleStruct struct {
			Field1 string `ytdlp:"field1"`
			Field2 int    `ytdlp:"field2"`
			Field3 bool   `ytdlp:"field3"`
		}

		expected := "Field1=%(field1|\"\")j|Field2=%(field2|0)j|Field3=%(field3|false)j"
		data, err := MarshalTemplate(SimpleStruct{Field1: "hello", Field2: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != expected {
			t.Errorf("expected %q but got %q", expected, data)
		}
	})

	t.Run("populated struct with missing tag", func(t *testing.T) {
		type MissingTagStruct struct {
			Field1 string `ytdlp:"field1"`
			Field2 int    `ytdlp:"field2"`
			Field3 bool
		}

		expected := "Field1=%(field1|\"\")j|Field2=%(field2|0)j|Field3=true"
		data, err := MarshalTemplate(MissingTagStruct{Field1: "hello", Field2: 100, Field3: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != expected {
			t.Errorf("expected %q but got %q", expected, data)
		}
	})

	t.Run("empty struct with missing tag", func(t *testing.T) {
		type MissingTagStruct struct {
			Field1 string `json:"field1"`
			Field2 int    `json:"field2"`
			Field3 bool   `json:"field3"`
		}

		expected := ""
		data, err := MarshalTemplate(MissingTagStruct{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != expected {
			t.Errorf("expected %q but got %q", expected, data)
		}
	})
}
