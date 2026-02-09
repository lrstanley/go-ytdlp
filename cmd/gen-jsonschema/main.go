// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/invopop/jsonschema"
	"github.com/lmittmann/tint"
	"github.com/lrstanley/go-ytdlp"
	"github.com/lrstanley/go-ytdlp/optiondata"
)

type UIDMapper struct {
	UID   string
	Props []string
}

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(
		os.Stderr,
		&tint.Options{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	)))

	ref := jsonschema.Reflector{
		AllowAdditionalProperties: false,
	}

	s := ref.Reflect(&ytdlp.FlagConfig{})

	for name, def := range s.Definitions {
		if def.Type != "object" {
			slog.Debug("skipping non-object definition", "name", name)
			continue
		}
		uid2prop := make(map[string][]string)

		for propPair := def.Properties.Oldest(); propPair != nil; propPair = propPair.Next() {
			uid, ok := propPair.Value.Extras["uid"]
			if !ok {
				continue
			}
			uid2prop[uid.(string)] = append(uid2prop[uid.(string)], propPair.Key)
		}
		slog.Debug("uid2prop", "uid2prop", uid2prop)

		// Convert uid2prop to an array, so we can sort and get a deterministic order, thus
		// preventing the jsonschema output from being different each time.
		uidPropMap := make([]UIDMapper, 0, len(uid2prop))
		for uid, props := range uid2prop {
			slices.Sort(props)
			uidPropMap = append(uidPropMap, UIDMapper{
				UID:   uid,
				Props: props,
			})
		}

		sort.Slice(uidPropMap, func(i, j int) bool {
			return uidPropMap[i].UID < uidPropMap[j].UID
		})

		for _, propMap := range uidPropMap {
			if len(propMap.Props) < 2 {
				continue
			}

			slog.Debug("adding all-of condition due to duplicate uids", "name", name, "props", propMap.Props)

			for _, prop := range propMap.Props {
				var disallowsOverride bool

				for _, opt := range optiondata.FindByID(prop) {
					if opt.NoOverride {
						disallowsOverride = true
						break
					}
				}

				if disallowsOverride {
					continue
				}

				slog.Debug("processing all-of for duplicate prop", "name", name, "prop", prop)

				var otherProps []string

				// Remove the current prop from the list of other props.
				for _, p := range propMap.Props {
					if p == prop {
						continue
					}
					otherProps = append(otherProps, p)
				}

				cond := &jsonschema.Schema{
					If: &jsonschema.Schema{Required: []string{prop}},
					Then: &jsonschema.Schema{
						Not: &jsonschema.Schema{Required: otherProps},
					},
				}

				def.AllOf = append(def.AllOf, cond)
			}
		}
	}

	f, err := os.OpenFile(filepath.Join(os.Args[1], "json-schema.json"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		slog.Error("failed to open file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(s)
	if err != nil {
		slog.Error("failed to marshal JSON", "error", err)
		os.Exit(1)
	}
}
