// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

// StderrCallbackFunc is a callback function that is called when stderr output
// is received from yt-dlp. Each invocation receives a single line of output,
// which may be terminated by either \n or \r (the latter is common for ffmpeg
// progress updates that overwrite in-place on a terminal).
type StderrCallbackFunc func(line string)

type stderrHandler struct {
	fn StderrCallbackFunc
}

func (h *stderrHandler) handle(line string) {
	if h == nil || h.fn == nil {
		return
	}
	h.fn(line)
}

// StderrFunc registers a callback function that will be called for every line
// of stderr output from yt-dlp. This includes ffmpeg progress lines that use
// \r (carriage return) for in-place terminal updates, which are not captured by
// [Command.ProgressFunc].
//
// Unlike [Command.ProgressFunc], no yt-dlp flags are injected - this is a pure
// listener on stderr output.
//
//   - See [Command.UnsetStderrFunc] for unsetting the stderr function.
func (c *Command) StderrFunc(fn StderrCallbackFunc) *Command {
	c.mu.Lock()
	c.stderr = &stderrHandler{fn: fn}
	c.mu.Unlock()
	return c
}

// UnsetStderrFunc removes the stderr callback function that was previously set
// with [Command.StderrFunc].
func (c *Command) UnsetStderrFunc() *Command {
	c.mu.Lock()
	c.stderr = nil
	c.mu.Unlock()
	return c
}
