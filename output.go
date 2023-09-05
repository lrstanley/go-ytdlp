// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"encoding/json"
	"slices"
	"sort"
	"time"
	"unicode"
)

type ResultLog struct {
	Timestamp time.Time      `json:"timestamp"`
	Line      string         `json:"line"`
	JSON      map[string]any `json:"json"` // May be nil if the log line wasn't valid JSON.
	Pipe      string         `json:"pipe"` // stdout or stderr.
}

type timestampWriter struct {
	checkJSON bool   // Whether to check if the log lines are valid JSON.
	pipe      string // stdout or stderr.

	buf            bytes.Buffer
	lastWriteStart time.Time
	results        []*ResultLog
}

func (w *timestampWriter) Write(p []byte) (n int, err error) {
	if w.lastWriteStart.IsZero() {
		w.lastWriteStart = time.Now()
	}

	if i := bytes.IndexByte(p, '\n'); i >= 0 {
		w.buf.Write(p[:i+1])
		w.flush()

		w.buf.Write(p[i+1:])

		return len(p), nil
	}

	return w.buf.Write(p)
}

func (w *timestampWriter) flush() {
	if w.buf.Len() == 0 {
		return
	}

	line := bytes.TrimRightFunc(w.buf.Bytes(), unicode.IsSpace)

	result := &ResultLog{
		Timestamp: w.lastWriteStart,
		Line:      string(line),
		Pipe:      w.pipe,
	}

	if w.checkJSON { // Try to parse the line as JSON.
		var jsonMap map[string]any
		if err := json.Unmarshal(line, &jsonMap); err == nil {
			result.JSON = jsonMap
		}
	}

	w.results = append(w.results, result)
	w.lastWriteStart = time.Time{}
	w.buf.Reset()
}

// mergeResults merges the results from this writer with the results from another writer
// (or multiple writers). The results are sorted by timestamp.
func (w *timestampWriter) mergeResults(otherWriters ...*timestampWriter) []*ResultLog {
	w.flush()

	results := slices.Clone(w.results)

	for _, other := range otherWriters {
		results = append(results, other.results...)
	}

	// Sort results by timestamp.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})

	return results
}

// String returns the contents of all log lines written to this writer.
func (w *timestampWriter) String() string {
	w.flush()

	var buf bytes.Buffer

	for i, r := range w.results {
		buf.WriteString(r.Line)

		if i < len(w.results)-1 {
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}
