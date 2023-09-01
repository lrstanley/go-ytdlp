// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
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
	OptionGroups []OptionGroup  `json:"option_groups"`
	Defaults     map[string]any `json:"defaults"`
}

func (c *CommandData) Generate() {
	for i := range c.OptionGroups {
		c.OptionGroups[i].Generate(c)
	}
}

type OptionGroup struct {
	// Generated fields.
	Name         string `json:"-"`
	NeedsStrconv bool   `json:"-"`

	// Command data fields.
	OriginalTitle string   `json:"title"`
	Description   string   `json:"description"`
	Options       []Option `json:"options"`
}

func (o *OptionGroup) Generate(data *CommandData) {
	o.Name = optionGroupReplacer.Replace(o.OriginalTitle)

	for i := range o.Options {
		o.Options[i].Generate(data)

		if !o.Options[i].IsExecutable && (o.Options[i].Type == "int" || o.Options[i].Type == "float64") {
			o.NeedsStrconv = true
		}
	}

	// Remove any ignored flags.
	o.Options = slices.DeleteFunc(o.Options, func(o Option) bool {
		return slices.Contains(ignoredFlags, o.Flag)
	})
}

type Option struct {
	// Generated fields.
	Name            string   `json:"-"` // simplified name, based off the first found flags.
	Flag            string   `json:"-"` // first flag (priority on long flags).
	AllFlags        []string `json:"-"` // all flags, short + long.
	MetaVarFuncArgs []string `json:"-"` // MetaVar converted to function arguments.
	IsExecutable    bool     `json:"-"` // if the option means yt-dlp doesn't accept arguments, and some callback is done.

	// Command data fields.
	Action  string   `json:"action"`
	Choices []string `json:"choices"`
	Default any      `json:"default"`
	Dest    string   `json:"dest"`
	Help    string   `json:"help"`
	MetaVar string   `json:"metavar"`
	Type    string   `json:"type"`
	Long    []string `json:"long"`
	Short   []string `json:"short"`
	Const   any      `json:"const"`
	NArgs   int      `json:"nargs"`
}

func (o *Option) Generate(data *CommandData) {
	o.AllFlags = append(o.Short, o.Long...) //nolint:gocritic

	if len(o.Long) > 0 {
		o.Name = strings.TrimPrefix(o.Long[0], "--")
		o.Flag = o.Long[0]
	} else if len(o.Short) > 0 {
		o.Name = strings.TrimPrefix(o.Short[0], "-")
		o.Flag = o.Short[0]
	}

	if o.Help == "SUPPRESSHELP" {
		o.Help = ""
	}

	if strings.Contains(o.Help, "%default") {
		if v, ok := data.Defaults[o.Dest]; ok {
			o.Help = strings.ReplaceAll(o.Help, "%default", fmt.Sprintf("%v", v))
		}
	}

	switch o.Type {
	case "", "None":
		switch o.Action {
		case "store_true", "store_false":
			o.Type = "bool"
		default:
			if o.NArgs > 0 {
				o.Type = "string"
			} else {
				o.Type = ""
			}
		}
	case "choice":
		o.Type = "string"
	case "float":
		o.Type = "float64"
	}

	if (o.Type == "" && o.NArgs < 1) ||
		strings.HasPrefix(o.Dest, "print_") ||
		strings.HasPrefix(o.Dest, "list_") ||
		strings.HasPrefix(o.Dest, "update_") ||
		strings.HasPrefix(o.Dest, "dump_") {
		o.IsExecutable = true
	}

	// Clean up [prefix:] syntax from MetaVar, since we don't care about the optional prefix type.
	re := regexp.MustCompile(`\[.*\]`)
	meta := re.ReplaceAllString(o.MetaVar, "")

	if slices.Contains(disallowedNames, meta) {
		meta = "value"
	}

	// Convert MetaVar to function arguments.
	for _, v := range strings.Split(meta, " ") {
		o.MetaVarFuncArgs = append(o.MetaVarFuncArgs, strcase.ToLowerCamel(strings.ToLower(v)))
	}

	// TODO: if choices, convert to ENUM, and make the enum type get passed in.
}
