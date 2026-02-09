// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/iancoleman/strcase"
	"github.com/lmittmann/tint"
)

var (
	funcMap = mergeFuncMaps(
		sprig.TxtFuncMap(), // http://masterminds.github.io/sprig/
		template.FuncMap{
			"last": func(x int, a interface{}) bool {
				return x == reflect.ValueOf(a).Len()-1 // if last index.
			},
			// https://github.com/iancoleman/strcase?tab=readme-ov-file#example
			"to_camel": func(s string) string {
				return acronymReplacer.Replace(strcase.ToCamel(s))
			}, // AnyKindOfString
			"to_lower_camel": func(s string) string {
				return acronymReplacer.Replace(strcase.ToLowerCamel(s))
			}, // anyKindOfString
			"to_snake": func(s string) string {
				return strcase.ToSnake(s)
			}, // any_kind_of_string
			"has_prefix": func(s, prefix string) bool {
				return strings.HasPrefix(s, prefix)
			},
			"has_suffix": func(s, suffix string) bool {
				return strings.HasSuffix(s, suffix)
			},
		},
	)

	constantsTmpl = template.Must(
		template.New("constants.gotmpl").
			Funcs(funcMap).
			ParseFiles("./templates/constants.gotmpl"),
	)

	builderTmpl = template.Must(
		template.New("builder.gotmpl").
			Funcs(funcMap).
			ParseGlob("./templates/builder*.gotmpl"),
	)

	builderTestTmpl = template.Must(
		template.New("buildertest.gotmpl").
			Funcs(funcMap).
			ParseGlob("./templates/builder*.gotmpl"),
	)

	commandJSONTmpl = template.Must(
		template.New("command_json.gen.gotmpl").
			Funcs(funcMap).
			ParseGlob("./templates/command_json*.gotmpl"),
	)

	optionDataTmpl = template.Must(
		template.New("optiondata.gotmpl").
			Funcs(funcMap).
			ParseGlob("./templates/optiondata*.gotmpl"),
	)

	optionGroupReplacer = strings.NewReplacer(
		" Options", "",
		" and ", " ",
	)

	// TODO: can be replaced when this is supported: https://github.com/iancoleman/strcase/issues/13
	acronymReplacer = strings.NewReplacer(
		"Api", "API",
		"Https", "HTTPS",
		"Http", "HTTP",
		"Id", "ID",
		"Json", "JSON",
		"Html", "HTML",
		"Xml", "XML",
		"Ascii", "ASCII",
		"Cpu", "CPU",
		"Dns", "DNS",
		"Ip", "IP",
		"Tls", "TLS",
		"Tcp", "TCP",
		"Ttl", "TTL",
		"Uuid", "UUID",
		"Uid", "UID",
		"Uri", "URI",
		"Url", "URL",
		"Xxs", "XXS",
		"Xff", "XFF",
		"Ffmpeg", "FFmpeg",
		"Avconv", "AVConv",
		"Mpegts", "MPEGTS",
		"mpegts", "mpegTS",
		"Mpeg", "MPEG",
		"Mpd", "MPD",
		"Mso", "MSO",
		"Cn", "CN",
		"Hls", "HLS",
		"Autonumber", "AutoNumber",
		"autonumber", "autoNumber",
		"Datebefore", "DateBefore",
		"Dateafter", "DateAfter",
		"datebefore", "dateBefore",
		"dateafter", "dateAfter",
		"Twofactor", "TwoFactor",
		"twofactor", "twoFactor",
		"Postprocessor", "PostProcessor",
		"postprocessor", "postProcessor",
		"Filesize", "FileSize",
		"filesize", "fileSize",
	)
)

func mergeFuncMaps(fm ...template.FuncMap) template.FuncMap {
	out := template.FuncMap{}
	for _, m := range fm {
		maps.Copy(out, m)
	}
	return out
}

func createTemplateFile(dir, name string, tmpl *template.Template, data any) {
	slog.Info("creating template file", "file", name)

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		panic(err)
	}

	name = filepath.Join(dir, name)

	// Check if the file exists first, and if it does, panic.
	if _, err = os.Stat(name); err == nil {
		panic(fmt.Sprintf("file %s already exists, not doing anything", name))
	}

	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		panic(err)
	}

	f.Close()
}

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(
		os.Stderr,
		&tint.Options{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	)))

	if len(os.Args) < 3 { //nolint:gomnd
		slog.Error("usage: codegen <command_data.json> <output_dir>")
		os.Exit(1)
	}

	var data OptionData

	slog.Info("reading option data file", "file", os.Args[1])
	optionDataFile, err := os.Open(os.Args[1])
	if err != nil {
		slog.Error("failed to open option data file", "error", err)
		os.Exit(1)
	}
	defer optionDataFile.Close()

	slog.Info("decoding option data")
	err = json.NewDecoder(optionDataFile).Decode(&data)
	if err != nil {
		slog.Error("failed to decode option data", "error", err)
		os.Exit(1)
	}

	data.Generate()

	createTemplateFile(os.Args[2], "optiondata/optiondata.gen.go", optionDataTmpl, data)
	createTemplateFile(os.Args[2], "constants.gen.go", constantsTmpl, data)
	createTemplateFile(os.Args[2], "builder.gen.go", builderTmpl, data)
	createTemplateFile(os.Args[2], "builder.gen_test.go", builderTestTmpl, data)
	createTemplateFile(os.Args[2], "command_json.gen.go", commandJSONTmpl, data)
}
