// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

// ignoredFlags are flags that are not intended to be used by the end-user and/or
// don't make sense in a binding library scenario.
//
//   - https://github.com/yt-dlp/yt-dlp/blob/master/README.md#developer-options
var ignoredFlags = []string{
	"--alias",                       // Higher-level abstraction, that go-ytdlp can do directly.
	"--allow-unplayable-formats",    // Not intended to be used by the end-user.
	"--export-options",              // Not intended to be used by the end-user.
	"--help",                        // Not needed.
	"--load-pages",                  // Not intended to be used by the end-user.
	"--no-allow-unplayable-formats", // Not intended to be used by the end-user.
	"--test",                        // Not intended to be used by the end-user.
	"--youtube-print-sig-code",      // Not intended to be used by the end-user.
}

// deprecatedFlags are flags that are deprecated (but still work), and should be replaced
// with alternatives (in most cases).
//
//   - https://github.com/yt-dlp/yt-dlp/blob/master/README.md#old-aliases
//   - https://github.com/yt-dlp/yt-dlp/blob/master/README.md#no-longer-supported
//   - https://github.com/yt-dlp/yt-dlp/blob/master/README.md#sponskrub-options
var deprecatedFlags = [][]string{
	{"--avconv-location", "Use [Command.FfmpegLocation] instead."},
	{"--call-home", "Not implemented."},
	{"--clean-infojson", "Use [Command.CleanInfoJson] instead."},
	{"--cn-verification-proxy", "Use [Command.GeoVerificationProxy] instead."},
	{"--dump-headers", "Use [Command.PrintTraffic] instead."},
	{"--dump-intermediate-pages", "Use [Command.DumpPages] instead."},
	{"--force-write-download-archive", "Use [Command.ForceWriteArchive] instead."},
	{"--include-ads", "No longer supported."},
	{"--load-info", "Use [Command.LoadInfoJson] instead."},
	{"--no-call-home", "This flag is now default in yt-dlp."},
	{"--no-clean-infojson", "Use [Command.NoCleanInfoJson] instead."},
	{"--no-include-ads", "This flag is now default in yt-dlp."},
	{"--no-split-tracks", "Use [Command.NoSplitChapters] instead."},
	{"--no-sponskrub-cut", "Use [Command.SponsorblockRemove] with \"-all\" as an argument."},
	{"--no-sponskrub-force", "No longer applicable."},
	{"--no-sponskrub", "Use [Command.NoSponsorblock] instead."},
	{"--no-write-annotations", "This flag is now default in yt-dlp."},
	{"--no-write-srt", "Use [Command.NoWriteSubs] instead."},
	{"--prefer-avconv", "avconv is not officially supported by yt-dlp."},
	{"--prefer-ffmpeg", "This flag is now default in yt-dlp."},
	{"--prefer-unsecure", "Use [Command.PreferInsecure] instead."},
	{"--rate-limit", "Use [Command.LimitRate] instead."},
	{"--split-tracks", "Use [Command.SplitChapters] instead."},
	{"--sponskrub-args", "No longer applicable."},
	{"--sponskrub-cut", "Use [Command.SponsorblockRemove] with \"all\" as an argument."},
	{"--sponskrub-force", "No longer applicable."},
	{"--sponskrub-location", "No longer applicable."},
	{"--sponskrub", "Use [Command.SponsorblockMark] with \"all\" as an argument."},
	{"--srt-lang", "Use [Command.SubLangs] instead."},
	{"--trim-file-names", "Use [Command.TrimFilenames] instead."},
	{"--write-annotations", "No supported site has annotations now."},
	{"--write-srt", "Use [Command.WriteSubs] instead."},
	{"--yes-overwrites", "Use [Command.ForceOverwrites] instead."},
}

var linkableFlags = map[string][]OptionURL{
	"--audio-multistreams": {{Name: "Format Selection", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#format-selection"}},
	"--compat-options":     {{Name: "Compatibility Options", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#differences-in-default-behavior"}},
	"--concat-playlist":    {{Name: "Output Template", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#output-template"}},
	"--dump-json":          {{Name: "Output Template", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#output-template"}},
	"--extractor-args":     {{Name: "Extractor Arguments", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#extractor-arguments"}},
	"--format-sort-force":  {{Name: "Sorting Formats", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#sorting-formats"}},
	"--format-sort": {
		{Name: "Sorting Formats", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#sorting-formats"},
		{Name: "Format Selection Examples", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#format-selection-examples"},
	},
	"--format": {
		{Name: "Format Selection", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#format-selection"},
		{Name: "Filter Formatting", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#filtering-formats"},
		{Name: "Format Selection Examples", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#format-selection-examples"},
	},
	"--output": {{Name: "Output Template", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#output-template"}},
	"--parse-metadata": {
		{Name: "Modifying Metadata", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#modifying-metadata"},
		{Name: "Modifying Metadata Examples", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#modifying-metadata-examples"},
	},
	"--replace-in-metadata": {
		{Name: "Modifying Metadata", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#modifying-metadata"},
		{Name: "Modifying Metadata Examples", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#modifying-metadata-examples"},
	},
	"--split-chapters":     {{Name: "Output Template", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#output-template"}},
	"--update-to":          {{Name: "Update Notes", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#update"}},
	"--update":             {{Name: "Update Notes", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#update"}},
	"--video-multistreams": {{Name: "Format Selection", URL: "https://github.com/yt-dlp/yt-dlp/blob/{version}/README.md#format-selection"}},
}

// knownExecutable are dest or flag names that are executable (override the default url input).
var knownExecutable = []string{
	"--update-to",
	"--update",
	"--version",
	"dump_user_agent",
	"list_extractor_descriptions",
	"list_extractors",
	"print_help",
}

var disallowedNames = []string{
	"",
	"type",
	"any",
	"str",
	"int",
	"int32",
	"int64",
	"float",
	"float32",
	"float64",
	"bool",
	"true",
	"false",
	"none",
}
