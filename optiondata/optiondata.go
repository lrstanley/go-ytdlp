// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

// Package optiondata contains the raw option data for go-ytdlp. Contents of this
// package are generated via cmd/codegen, and may change at any time.
package optiondata

// OptionGroup is a group of options (e.g. general, verbosity, etc).
type OptionGroup struct {
	// Name of the option group.
	Name string `json:"name"`
	// Description of the option group, if any.
	Description string `json:"description,omitempty"`
	// Options are the options within the group.
	Options []*Option `json:"options"`
}

// Option is the raw option data for the given option (flag, essentially).
type Option struct {
	// ID is the identifier for the option, if one exists (may not for executables).
	// Note that this ID is not unique, as multiple options can have the same ID
	// (e.g. --something and --no-something).
	ID string `json:"id,omitempty"`
	// Name is the simplified name, based off the first found flags.
	Name string `json:"name"`
	// NameCamelCase is the same as [Option.Name], but in camelCase.
	NameCamelCase string `json:"name_camel_case"`
	// NamePascalCase is the same as [Option.Name], but in PascalCase.
	NamePascalCase string `json:"name_pascal_case"`
	// Links are optional links to the documentation for the option.
	URLs []*OptionURL `json:"urls,omitempty"`
	// DefaultFlag is the first flag (priority on long flags).
	DefaultFlag string `json:"default_flag"`
	// ArgNames are the argument names, if any -- length should match [Option.NArgs].
	ArgNames []string `json:"arg_names,omitempty"`
	// Executable is true if the option doesn't accept arguments.
	Executable bool `json:"executable"`
	// Deprecated will contain the deprecation description if the option if deprecated.
	Deprecated string `json:"deprecated,omitempty"`
	// Choices contains the list of required inputs for the option, if the option
	// has restricted inputs.
	Choices []string `json:"choices"`
	// Help contains the help text for the option.
	Help string `json:"help,omitempty"`
	// Hidden is true if the option is not returned in the help output (but can
	// still be provided).
	Hidden bool `json:"hidden"`
	// MetaArgs are the simplified syntax for the option, if any.
	MetaArgs string `json:"meta_args,omitempty"`
	// Type is the type (string, int, float64, bool, etc) of the option.
	Type string `json:"type"`
	// LongFlags are the extended flags for the option (e.g. --version).
	LongFlags []string `json:"long_flags"`
	// ShortFlags are the shortened flags for the option (e.g. -v).
	ShortFlags []string `json:"short_flags"`
	// NArgs is the number of arguments the option accepts.
	NArgs int `json:"nargs"`
}

type OptionURL struct {
	// Name is the name of the option link.
	Name string `json:"name"`
	// URL is the link to the documentation for the option.
	URL string `json:"url"`
}
