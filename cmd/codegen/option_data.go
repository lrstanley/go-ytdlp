// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"regexp"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
)

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

type OptionURL struct {
	Name string
	URL  string
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

type OptionData struct {
	Channel      string        `json:"channel"`
	Version      string        `json:"version"`
	OptionGroups []OptionGroup `json:"option_groups"`
}

func (c *OptionData) Generate() {
	for i := range c.OptionGroups {
		c.OptionGroups[i].Generate(c)
	}
}

type OptionGroup struct {
	// Generated fields.
	Parent *OptionData `json:"-"` // Reference to parent.
	Name   string      `json:"-"`

	// Command data fields.
	OriginalName string   `json:"name"`
	Description  string   `json:"description"`
	Options      []Option `json:"options"`
}

func (o *OptionGroup) Generate(parent *OptionData) {
	o.Parent = parent
	o.Name = optionGroupReplacer.Replace(o.OriginalName)

	for i := range o.Options {
		o.Options[i].Generate(o)
	}

	// Remove any ignored flags.
	o.Options = slices.DeleteFunc(o.Options, func(o Option) bool {
		return slices.Contains(ignoredFlags, o.Flag)
	})
}

type Option struct {
	// Generated fields.
	Parent     *OptionGroup `json:"-"` // Reference to parent.
	Name       string       `json:"-"` // simplified name, based off the first found flags.
	Flag       string       `json:"-"` // first flag (priority on long flags).
	AllFlags   []string     `json:"-"` // all flags, short + long.
	ArgNames   []string     `json:"-"` // MetaArgs converted to function arguments.
	Executable bool         `json:"-"` // if the option means yt-dlp doesn't accept arguments, and some callback is done.
	Deprecated string       `json:"-"` // if the option is deprecated, this will be the deprecation description.
	URLs       []OptionURL  `json:"-"` // if the option has any links to the documentation.

	// Command data fields.
	ID           string   `json:"id"`
	Action       string   `json:"action"`
	Choices      []string `json:"choices"`
	Help         string   `json:"help"`
	Hidden       bool     `json:"hidden"`
	MetaArgs     string   `json:"meta_args"`
	Type         string   `json:"type"`
	LongFlags    []string `json:"long_flags"`
	ShortFlags   []string `json:"short_flags"`
	NArgs        int      `json:"nargs"`
	DefaultValue any      `json:"default_value"`
	Const        any      `json:"const_value"`
}

var (
	reMetaArgsStrip = regexp.MustCompile(`\[.*\]`)
	reRemoveAlias   = regexp.MustCompile(`\s+\(Alias:.*\)`)
)

func (o *Option) Generate(parent *OptionGroup) {
	o.Parent = parent
	o.AllFlags = append(o.ShortFlags, o.LongFlags...) //nolint:gocritic

	if len(o.LongFlags) > 0 {
		o.Name = strings.TrimPrefix(o.LongFlags[0], "--")
		o.Flag = o.LongFlags[0]
	} else if len(o.ShortFlags) > 0 {
		o.Name = strings.TrimPrefix(o.ShortFlags[0], "-")
		o.Flag = o.ShortFlags[0]
	}

	if slices.Contains(knownExecutable, o.ID) || slices.Contains(knownExecutable, o.Flag) {
		o.Executable = true
	}

	for _, d := range deprecatedFlags {
		if strings.EqualFold(d[0], o.ID) || strings.EqualFold(d[0], o.Flag) {
			o.Deprecated = d[1]
		}
	}

	switch o.Type {
	case "choice":
		o.Type = "string"
	case "float":
		o.Type = "float64"
	case "":
		if o.NArgs == 0 {
			o.Type = "bool"
		} else {
			o.Type = "string"
		}
	}

	// Clean up help text.
	o.Help = reRemoveAlias.ReplaceAllString(o.Help, "")

	// Clean up [prefix:] syntax from MetaArgs, since we don't care about the optional prefix type.
	meta := reMetaArgsStrip.ReplaceAllString(o.MetaArgs, "")

	if slices.Contains(disallowedNames, meta) {
		meta = "value"
	}

	// Convert MetaArgs to function arguments.
	for _, v := range strings.Split(meta, " ") {
		o.ArgNames = append(o.ArgNames, strcase.ToLowerCamel(strings.ToLower(v)))
	}

	// URLs.
	if urls, ok := linkableFlags[o.Flag]; ok {
		for _, u := range urls {
			o.URLs = append(o.URLs, OptionURL{
				Name: u.Name,
				URL:  strings.ReplaceAll(u.URL, "{version}", parent.Parent.Version),
			})
		}
	}
}
