// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLogSinkOrderAndNotConcurrent(t *testing.T) {
	t.Parallel()

	var inCallback atomic.Bool
	var mu sync.Mutex
	var lines []string

	sink := &logSink{fn: func(log *ResultLog) {
		if !inCallback.CompareAndSwap(false, true) {
			t.Error("LogFunc called concurrently")
		}
		time.Sleep(5 * time.Millisecond)
		mu.Lock()
		lines = append(lines, string(log.Pipe)+":"+log.Line)
		mu.Unlock()
		inCallback.Store(false)
	}}

	stdout := &timestampWriter{pipe: PipeStdout, log: sink}
	stderr := &timestampWriter{pipe: PipeStderr, log: sink}

	if _, err := stdout.Write([]byte("out-1\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := stderr.Write([]byte("err-1\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := stdout.Write([]byte("out-2\n")); err != nil {
		t.Fatal(err)
	}

	want := []string{"stdout:out-1", "stderr:err-1", "stdout:out-2"}
	mu.Lock()
	defer mu.Unlock()
	if len(lines) != len(want) {
		t.Fatalf("lines = %v, want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("lines = %v, want %v", lines, want)
		}
	}
}

func TestLogSinkNotConcurrentAcrossPipes(t *testing.T) {
	t.Parallel()

	var inCallback atomic.Bool
	var overlap atomic.Bool

	sink := &logSink{fn: func(_ *ResultLog) {
		if !inCallback.CompareAndSwap(false, true) {
			overlap.Store(true)
		}
		time.Sleep(20 * time.Millisecond)
		inCallback.Store(false)
	}}

	stdout := &timestampWriter{pipe: PipeStdout, log: sink}
	stderr := &timestampWriter{pipe: PipeStderr, log: sink}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = stdout.Write([]byte("a\n"))
	}()
	go func() {
		defer wg.Done()
		_, _ = stderr.Write([]byte("b\n"))
	}()
	wg.Wait()

	if overlap.Load() {
		t.Fatal("LogFunc overlapped across pipes")
	}
}

func TestFlushCRLogsButNotResults(t *testing.T) {
	t.Parallel()

	var got []*ResultLog
	sink := &logSink{fn: func(log *ResultLog) {
		got = append(got, log)
	}}
	w := &timestampWriter{pipe: PipeStderr, log: sink}

	if _, err := w.Write([]byte("frame1\rframe2\r")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("done\n")); err != nil {
		t.Fatal(err)
	}

	if len(got) != 3 {
		t.Fatalf("logs = %d, want 3", len(got))
	}
	if !got[0].Overwrite || got[0].Line != "frame1" || got[0].Pipe != PipeStderr {
		t.Fatalf("got[0] = %+v", got[0])
	}
	if !got[1].Overwrite || got[1].Line != "frame2" {
		t.Fatalf("got[1] = %+v", got[1])
	}
	if got[2].Overwrite || got[2].Line != "done" {
		t.Fatalf("got[2] = %+v", got[2])
	}
	if w.String() != "done" {
		t.Fatalf("String() = %q, want done", w.String())
	}
}

func TestPipeUntypedStringCompare(t *testing.T) {
	t.Parallel()

	log := ResultLog{Pipe: PipeStderr}
	if log.Pipe != "stderr" {
		t.Fatalf("Pipe = %q, want stderr", log.Pipe)
	}
	if log.Pipe != PipeStderr {
		t.Fatalf("Pipe = %q, want PipeStderr", log.Pipe)
	}
	log.Pipe = PipeStdout
	if log.Pipe != "stdout" {
		t.Fatalf("Pipe = %q, want stdout", log.Pipe)
	}
}
