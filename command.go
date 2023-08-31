// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"errors"
)

// addFlag adds a flag to the command. If a flag with the same ID already
// exists, it will be replaced.
func (c *Command) addFlag(f *Flag) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, ff := range c.flags {
		if ff.ID == f.ID {
			c.flags[i] = f
			return
		}
	}

	c.flags = append(c.flags, f)
}

// removeFlagByID removes a flag from the command by its ID.
func (c *Command) removeFlagByID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, f := range c.flags {
		if f.ID == id {
			c.flags = append(c.flags[:i], c.flags[i+1:]...)
			return
		}
	}
}

func (c *Command) Run(args ...string) (*Results, error) {
	// TODO
	return nil, errors.New("todo")
}

type Results struct{}

type Flag struct {
	ID   string // Unique ID to ensure boolean flags are not duplicated.
	Flag string // Actual flag, e.g. "--version".

	Args []string // Optional args. If nil, it's a boolean flag.
}
