// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.
//
// Code generated by cmd/codegen. DO NOT EDIT.
//
// Subtitle Option Group

package ytdlp

// Write subtitle file
//
//   - See [Command.UnsetWriteSubs], for unsetting the flag.
//   - WriteSubs maps to cli flags: --write-subs/--write-srt.
func (c *Command) WriteSubs() *Command {
	c.addFlag(&Flag{
		ID:   "writesubtitles",
		Flag: "--write-subs",
		Args: nil,
	})
	return c
}

// UnsetWriteSubs unsets any flags that were previously set by
// [WriteSubs].
func (c *Command) UnsetWriteSubs() *Command {
	c.removeFlagByID("writesubtitles")
	return c
}

// Do not write subtitle file (default)
//
//   - See [Command.UnsetWriteSubs], for unsetting the flag.
//   - NoWriteSubs maps to cli flags: --no-write-subs/--no-write-srt.
func (c *Command) NoWriteSubs() *Command {
	c.addFlag(&Flag{
		ID:   "writesubtitles",
		Flag: "--no-write-subs",
		Args: nil,
	})
	return c
}

// Write automatically generated subtitle file
//
//   - See [Command.UnsetWriteAutoSubs], for unsetting the flag.
//   - WriteAutoSubs maps to cli flags: --write-auto-subs/--write-automatic-subs.
func (c *Command) WriteAutoSubs() *Command {
	c.addFlag(&Flag{
		ID:   "writeautomaticsub",
		Flag: "--write-auto-subs",
		Args: nil,
	})
	return c
}

// UnsetWriteAutoSubs unsets any flags that were previously set by
// [WriteAutoSubs].
func (c *Command) UnsetWriteAutoSubs() *Command {
	c.removeFlagByID("writeautomaticsub")
	return c
}

// Do not write auto-generated subtitles (default)
//
//   - See [Command.UnsetWriteAutoSubs], for unsetting the flag.
//   - NoWriteAutoSubs maps to cli flags: --no-write-auto-subs/--no-write-automatic-subs.
func (c *Command) NoWriteAutoSubs() *Command {
	c.addFlag(&Flag{
		ID:   "writeautomaticsub",
		Flag: "--no-write-auto-subs",
		Args: nil,
	})
	return c
}

// AllSubs sets the "all-subs" flag (no description specified).
//
//   - See [Command.UnsetAllSubs], for unsetting the flag.
//   - AllSubs maps to cli flags: --all-subs (hidden).
func (c *Command) AllSubs() *Command {
	c.addFlag(&Flag{
		ID:   "allsubtitles",
		Flag: "--all-subs",
		Args: nil,
	})
	return c
}

// UnsetAllSubs unsets any flags that were previously set by
// [AllSubs].
func (c *Command) UnsetAllSubs() *Command {
	c.removeFlagByID("allsubtitles")
	return c
}

// List available subtitles of each video. Simulate unless --no-simulate is used
//
//   - See [Command.UnsetListSubs], for unsetting the flag.
//   - ListSubs maps to cli flags: --list-subs.
func (c *Command) ListSubs() *Command {
	c.addFlag(&Flag{
		ID:   "listsubtitles",
		Flag: "--list-subs",
		Args: nil,
	})
	return c
}

// UnsetListSubs unsets any flags that were previously set by
// [ListSubs].
func (c *Command) UnsetListSubs() *Command {
	c.removeFlagByID("listsubtitles")
	return c
}

// Subtitle format; accepts formats preference, e.g. "srt" or "ass/srt/best"
//
//   - See [Command.UnsetSubFormat], for unsetting the flag.
//   - SubFormat maps to cli flags: --sub-format=FORMAT.
func (c *Command) SubFormat(format string) *Command {
	c.addFlag(&Flag{
		ID:   "subtitlesformat",
		Flag: "--sub-format",
		Args: []string{format},
	})
	return c
}

// UnsetSubFormat unsets any flags that were previously set by
// [SubFormat].
func (c *Command) UnsetSubFormat() *Command {
	c.removeFlagByID("subtitlesformat")
	return c
}

// Languages of the subtitles to download (can be regex) or "all" separated by
// commas, e.g. --sub-langs "en.*,ja". You can prefix the language code with a "-"
// to exclude it from the requested languages, e.g. --sub-langs all,-live_chat. Use
// --list-subs for a list of available language tags
//
//   - See [Command.UnsetSubLangs], for unsetting the flag.
//   - SubLangs maps to cli flags: --sub-langs/--srt-langs=LANGS.
func (c *Command) SubLangs(langs string) *Command {
	c.addFlag(&Flag{
		ID:   "subtitleslangs",
		Flag: "--sub-langs",
		Args: []string{langs},
	})
	return c
}

// UnsetSubLangs unsets any flags that were previously set by
// [SubLangs].
func (c *Command) UnsetSubLangs() *Command {
	c.removeFlagByID("subtitleslangs")
	return c
}
