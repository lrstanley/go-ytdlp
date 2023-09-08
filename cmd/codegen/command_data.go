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

var ignoredFlags = []string{
	"--help",
	"--export-options",
	"--alias",
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
