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
	{"--all-formats", "Use [Command.Format] with `all` as an argument."},
	{"--all-subs", "Use [Command.SubLangs] with `all` as an argument, in addition to [Command.WriteSubs]."},
	{"--autonumber-size", "Use string formatting, e.g. `%(autonumber)03d`."},
	{"--autonumber-start", "Use internal field formatting like `%(autonumber+NUMBER)s`."},
	{"--avconv-location", "Use [Command.FfmpegLocation] instead."},
	{"--break-on-reject", "Use [Command.BreakMatchFilters] instead."},
	{"--call-home", "Not implemented."},
	{"--clean-infojson", "Use [Command.CleanInfoJson] instead."},
	{"--cn-verification-proxy", "Use [Command.GeoVerificationProxy] instead."},
	{"--dump-headers", "Use [Command.PrintTraffic] instead."},
	{"--dump-intermediate-pages", "Use [Command.DumpPages] instead."},
	{"--exec-before-download", "Use [Command.Exec] with `before_dl:CMD` as an argument."},
	{"--force-generic-extractor", "Use [Command.UseExtractors] with `generic,default` as an argument."},
	{"--force-write-download-archive", "Use [Command.ForceWriteArchive] instead."},
	{"--geo-bypass-country", "Use [Command.XFF] with `CODE` as an argument."},
	{"--geo-bypass-ip-block", "Use [Command.XFF] with `IP_BLOCK` as an argument."},
	{"--geo-bypass", "Use [Command.XFF] with `default` as an argument."},
	{"--get-description", "Use [Command.Print] with `description` as an argument."},
	{"--get-duration", "Use [Command.Print] with `duration_string` as an argument."},
	{"--get-filename", "Use [Command.Print] with `filename` as an argument."},
	{"--get-format", "Use [Command.Print] with `format` as an argument."},
	{"--get-id", "Use [Command.Print] with `id` as an argument."},
	{"--get-thumbnail", "Use [Command.Print] with `thumbnail` as an argument."},
	{"--get-title", "Use [Command.Print] with `title` as an argument."},
	{"--get-url", "Use [Command.Print] with `urls` as an argument."},
	{"--hls-prefer-ffmpeg", "Use [Command.Downloader] with `m3u8:ffmpeg` as an argument."},
	{"--hls-prefer-native", "Use [Command.Downloader] with `m3u8:native` as an argument."},
	{"--id", "Use [Command.Output] with `%(id)s.%(ext)s` as an argument."},
	{"--include-ads", "No longer supported."},
	{"--list-formats-as-table", "Use [Command.ListFormatsAsTable] or [Command.CompatOptions] with `-list-formats` as an argument."},
	{"--list-formats-old", "Use [Command.CompatOptions] with `list-formats` as an argument."},
	{"--list-formats", "Use [Command.Print] with `formats_table` as an argument."},
	{"--list-thumbnails", "Call [Command.Print] twice, once with `thumbnails_table` as an argument, then with `playlist:thumbnails_table` as an argument."},
	{"--load-info", "Use [Command.LoadInfoJson] instead."},
	{"--match-title", "Use [Command.MatchFilters] instead (e.g. `title ~= (?i)REGEX`)."},
	{"--max-views", "Use [Command.MatchFilters] instead (e.g. `view_count <=? COUNT`)."},
	{"--metadata-from-title", "Use [Command.ParseMetadata] with `%(title)s:FORMAT` as an argument."},
	{"--min-views", "Use [Command.MatchFilters] instead (e.g. `view_count >=? COUNT`)."},
	{"--no-call-home", "This flag is now default in yt-dlp."},
	{"--no-clean-infojson", "Use [Command.NoCleanInfoJson] instead."},
	{"--no-colors", "Use [Command.Color] with `no_color` as an argument."},
	{"--no-exec-before-download", "Use [Command.NoExec] instead."},
	{"--no-geo-bypass", "Use [Command.XFF] with `never` as an argument."},
	{"--no-include-ads", "This flag is now default in yt-dlp."},
	{"--no-playlist-reverse", "It is now the default behavior."},
	{"--no-split-tracks", "Use [Command.NoSplitChapters] instead."},
	{"--no-sponskrub-cut", "Use [Command.SponsorblockRemove] with `-all` as an argument."},
	{"--no-sponskrub-force", "No longer applicable."},
	{"--no-sponskrub", "Use [Command.NoSponsorblock] instead."},
	{"--no-write-annotations", "This flag is now default in yt-dlp."},
	{"--no-write-srt", "Use [Command.NoWriteSubs] instead."},
	{"--playlist-end", "Use [Command.PlaylistItems] with `:<your-number>` as an argument."},
	{"--playlist-reverse", "Use [Command.PlaylistItems] with `::-1` as an argument."},
	{"--playlist-start", "Use [Command.PlaylistItems] with `<your-number>:` as an argument."},
	{"--prefer-avconv", "avconv is not officially supported by yt-dlp."},
	{"--prefer-ffmpeg", "This flag is now default in yt-dlp."},
	{"--prefer-unsecure", "Use [Command.PreferInsecure] instead."},
	{"--rate-limit", "Use [Command.LimitRate] instead."},
	{"--referer", "Use [Command.AddHeaders] instead (e.g. `Referer:URL`)."},
	{"--reject-title", "Use [Command.MatchFilters] instead (e.g. `title !~= (?i)REGEX`)."},
	{"--split-tracks", "Use [Command.SplitChapters] instead."},
	{"--sponskrub-args", "No longer applicable."},
	{"--sponskrub-cut", "Use [Command.SponsorblockRemove] with `all` as an argument."},
	{"--sponskrub-force", "No longer applicable."},
	{"--sponskrub-location", "No longer applicable."},
	{"--sponskrub", "Use [Command.SponsorblockMark] with `all` as an argument."},
	{"--srt-lang", "Use [Command.SubLangs] instead."},
	{"--trim-file-names", "Use [Command.TrimFilenames] instead."},
	{"--user-agent", "Use [Command.AddHeaders] instead (e.g. `User-Agent:UA`)."},
	{"--write-annotations", "No supported site has annotations now."},
	{"--write-srt", "Use [Command.WriteSubs] instead."},
	{"--yes-overwrites", "Use [Command.ForceOverwrites] instead."},
	{"--youtube-include-dash-manifest", "Use [Command.YoutubeIncludeDashManifest] instead."},
	{"--youtube-include-hls-manifest", "Use [Command.YoutubeIncludeHLSManifest] instead."},
	{"--youtube-skip-dash-manifest", "Use [Command.ExtractorArgs] with `youtube:skip=dash` as an argument."},
	{"--youtube-skip-hls-manifest", "Use [Command.ExtractorArgs] with `youtube:skip=hls` as an argument."},
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
