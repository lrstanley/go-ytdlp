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
	{"--no-write-annotations", "This flag is now default in yt-dlp."},
	{"--no-write-srt", "Use [Command.NoWriteSubs] instead."},
	{"--prefer-avconv", "avconv is not officially supported by yt-dlp."},
	{"--prefer-ffmpeg", "This flag is now default in yt-dlp."},
	{"--prefer-unsecure", "Use [Command.PreferInsecure] instead."},
	{"--rate-limit", "Use [Command.LimitRate] instead."},
	{"--split-tracks", "Use [Command.SplitChapters] instead."},
	{"--srt-lang", "Use [Command.SubLangs] instead."},
	{"--trim-file-names", "Use [Command.TrimFilenames] instead."},
	{"--write-annotations", "No supported site has annotations now."},
	{"--write-srt", "Use [Command.WriteSubs] instead."},
	{"--yes-overwrites", "Use [Command.ForceOverwrites] instead."},
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

type CommandData struct {
	Channel      string        `json:"channel"`
	Version      string        `json:"version"`
	OptionGroups []OptionGroup `json:"option_groups"`
}

func (c *CommandData) Generate() {
	for i := range c.OptionGroups {
		c.OptionGroups[i].Generate(c)
	}
}

type OptionGroup struct {
	// Generated fields.
	Parent *CommandData `json:"-"` // Reference to parent.
	Name   string       `json:"-"`

	// Command data fields.
	OriginalTitle string   `json:"title"`
	Description   string   `json:"description"`
	Options       []Option `json:"options"`
}

func (o *OptionGroup) Generate(parent *CommandData) {
	o.Parent = parent
	o.Name = optionGroupReplacer.Replace(o.OriginalTitle)

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
	Parent          *OptionGroup `json:"-"` // Reference to parent.
	Name            string       `json:"-"` // simplified name, based off the first found flags.
	Flag            string       `json:"-"` // first flag (priority on long flags).
	AllFlags        []string     `json:"-"` // all flags, short + long.
	MetaVarFuncArgs []string     `json:"-"` // MetaVar converted to function arguments.
	IsExecutable    bool         `json:"-"` // if the option means yt-dlp doesn't accept arguments, and some callback is done.
	Deprecated      string       `json:"-"` // if the option is deprecated, this will be the deprecation description.

	// Command data fields.
	Action  string   `json:"action"`
	Choices []string `json:"choices"`
	Default any      `json:"default"`
	Dest    string   `json:"dest"`
	Help    string   `json:"help"`
	Hidden  bool     `json:"hidden"`
	MetaVar string   `json:"metavar"`
	Type    string   `json:"type"`
	Long    []string `json:"long"`
	Short   []string `json:"short"`
	Const   any      `json:"const"`
	NArgs   int      `json:"nargs"`
}

var (
	reMetaVarStrip = regexp.MustCompile(`\[.*\]`)
	reRemoveAlias  = regexp.MustCompile(`\s+\(Alias:.*\)`)
)

func (o *Option) Generate(parent *OptionGroup) {
	o.Parent = parent
	o.AllFlags = append(o.Short, o.Long...) //nolint:gocritic

	if len(o.Long) > 0 {
		o.Name = strings.TrimPrefix(o.Long[0], "--")
		o.Flag = o.Long[0]
	} else if len(o.Short) > 0 {
		o.Name = strings.TrimPrefix(o.Short[0], "-")
		o.Flag = o.Short[0]
	}

	if slices.Contains(knownExecutable, o.Dest) || slices.Contains(knownExecutable, o.Flag) {
		o.IsExecutable = true
	}

	for _, d := range deprecatedFlags {
		if strings.EqualFold(d[0], o.Dest) || strings.EqualFold(d[0], o.Flag) {
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

	// Clean up [prefix:] syntax from MetaVar, since we don't care about the optional prefix type.
	meta := reMetaVarStrip.ReplaceAllString(o.MetaVar, "")

	if slices.Contains(disallowedNames, meta) {
		meta = "value"
	}

	// Convert MetaVar to function arguments.
	for _, v := range strings.Split(meta, " ") {
		o.MetaVarFuncArgs = append(o.MetaVarFuncArgs, strcase.ToLowerCamel(strings.ToLower(v)))
	}
}
