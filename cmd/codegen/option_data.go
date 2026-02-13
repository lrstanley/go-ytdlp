// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
)

type Extractor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AgeLimit    int    `json:"age_limit"`
}

type OptionURL struct {
	Name string
	URL  string
}

type OptionData struct {
	Channel      string               `json:"channel"`
	Version      string               `json:"version"`
	OptionGroups []OptionGroup        `json:"option_groups"`
	OptionIDs    map[string][]*Option `json:"option_ids"`
	Extractors   []Extractor          `json:"extractors"`
}

func (c *OptionData) Generate() {
	for i := range c.OptionGroups {
		c.OptionGroups[i].Generate(c)
		slog.Info("generated option group", "group", c.OptionGroups[i].Name)
	}

	for _, g := range c.OptionGroups {
		for _, o := range g.Options {
			c.OptionIDs[o.ID] = append(c.OptionIDs[o.ID], &o)
		}
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
		slog.Info("generated option", "option", o.Options[i].Flag)
	}

	// Remove any ignored flags.
	o.Options = slices.DeleteFunc(o.Options, func(o Option) bool {
		return slices.Contains(ignoredFlags, o.Flag)
	})
}

func (o *OptionGroup) AllAllowsMultiple() (opts []*Option) {
	for _, o := range o.Options {
		if o.AllowsMultiple {
			opts = append(opts, &o)
		}
	}
	return opts
}

type Option struct {
	// Generated fields.
	Parent         *OptionGroup `json:"-"` // Reference to parent.
	Name           string       `json:"-"` // simplified name, based off the first found flags.
	Flag           string       `json:"-"` // first flag (priority on long flags).
	AllFlags       []string     `json:"-"` // all flags, short + long.
	ArgNames       []string     `json:"-"` // MetaArgs converted to function arguments.
	Executable     bool         `json:"-"` // if the option means yt-dlp doesn't accept arguments, and some callback is done.
	NoOverride     bool         `json:"-"` // if the option should not override other flags with the same ID.
	Deprecated     string       `json:"-"` // if the option is deprecated, this will be the deprecation description.
	URLs           []OptionURL  `json:"-"` // if the option has any links to the documentation.
	AllowsMultiple bool         `json:"-"` // if the option allows being invoked multiple times.

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

	if strings.Contains(o.Help, "used multiple times") || strings.Contains(o.Help, "option multiple times") {
		o.AllowsMultiple = true
	}

	for _, d := range deprecatedFlags {
		if strings.EqualFold(d[0], o.ID) || strings.EqualFold(d[0], o.Flag) {
			o.Deprecated = d[1]
		}
	}

	if slices.Contains(noOverrideIDs, o.ID) {
		o.NoOverride = true
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
