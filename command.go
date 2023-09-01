// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

func New() *Command {
	cmd := &Command{
		env: make(map[string]string),
	}

	return cmd
}

type Command struct {
	mu         sync.RWMutex
	executable string
	directory  string
	env        map[string]string
	flags      []*Flag
}

func (c *Command) Clone() *Command {
	c.mu.RLock()
	cc := &Command{
		executable: c.executable,
		directory:  c.directory,
		env:        make(map[string]string, len(c.env)),
		flags:      make([]*Flag, len(c.flags)),
	}

	for k, v := range c.env {
		cc.env[k] = v
	}

	for i, f := range c.flags {
		cc.flags[i] = f.Clone()
	}
	c.mu.RUnlock()

	return cc
}

// SetExecutable sets the executable path to yt-dlp for the command.
func (c *Command) SetExecutable(path string) {
	c.mu.Lock()
	c.executable = path
	c.mu.Unlock()
}

// SetWorkDir sets the working directory for the command.
func (c *Command) SetWorkDir(path string) {
	c.mu.Lock()
	c.directory = path
	c.mu.Unlock()
}

func (c *Command) SetEnvVar(key, value string) {
	c.mu.Lock()
	if value == "" {
		delete(c.env, key)
	} else {
		c.env[key] = value
	}
	c.mu.Unlock()
}

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

func (c *Command) Run(ctx context.Context, args ...string) *Results {
	var cmdArgs []string

	for _, f := range c.flags {
		cmdArgs = append(cmdArgs, f.Raw()...)
	}

	cmdArgs = append(cmdArgs, args...) // URLs or similar.

	var name string

	c.mu.RLock()
	name = c.executable

	if name == "" {
		if runtime.GOOS == "windows" {
			name = "yt-dlp.exe"
		} else {
			name = "yt-dlp"
		}
	}

	cmd := exec.CommandContext(ctx, name, cmdArgs...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if c.directory != "" {
		cmd.Dir = c.directory
	}

	if len(c.env) > 0 {
		cmd.Env = make([]string, 0, len(c.env))
		for k, v := range c.env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	c.mu.RUnlock()

	err := cmd.Run()

	return &Results{
		Executable: cmd.Path,
		Args:       cmd.Args[1:],
		ExitCode:   cmd.ProcessState.ExitCode(),
		Stdout:     stdout.Bytes(),
		Stderr:     stderr.Bytes(),
		Error:      err,
	}
}

type Results struct {
	Executable string
	Args       []string
	ExitCode   int
	Stdout     []byte
	Stderr     []byte

	Error error
}

type Flag struct {
	ID   string // Unique ID to ensure boolean flags are not duplicated.
	Flag string // Actual flag, e.g. "--version".

	Args []string // Optional args. If nil, it's a boolean flag.
}

func (f *Flag) Clone() *Flag {
	return &Flag{
		ID:   f.ID,
		Flag: f.Flag,
		Args: f.Args,
	}
}

func (f *Flag) Raw() []string {
	if f.Args == nil {
		return []string{f.Flag}
	}

	return append([]string{f.Flag}, strings.Join(f.Args, " "))
}
