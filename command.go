// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"context"
	"os/exec"
	"runtime"
	"slices"
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
func (c *Command) SetExecutable(path string) *Command {
	c.mu.Lock()
	c.executable = path
	c.mu.Unlock()

	return c
}

// SetWorkDir sets the working directory for the command. Defaults to current working
// directory.
func (c *Command) SetWorkDir(path string) *Command {
	c.mu.Lock()
	c.directory = path
	c.mu.Unlock()

	return c
}

// SetEnvVar sets an environment variable for the command. If value is empty, it will
// be removed.
func (c *Command) SetEnvVar(key, value string) *Command {
	c.mu.Lock()
	if value == "" {
		delete(c.env, key)
	} else {
		c.env[key] = value
	}
	c.mu.Unlock()

	return c
}

// getFlagsByID returns all flags with the provided ID/"dest".
func (c *Command) getFlagsByID(id string) []*Flag {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var flags []*Flag

	for _, f := range c.flags {
		if f.ID == id {
			flags = append(flags, f)
		}
	}

	return flags
}

// addFlag adds a flag to the command. If a flag with the same ID/"dest" already
// exists, it will be replaced.
func (c *Command) addFlag(f *Flag) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If boolean flag, ensure it's not duplicated.
	if f.Args == nil {
		for i, ff := range c.flags {
			if ff.ID == f.ID {
				c.flags[i] = f
				return
			}
		}
	}

	c.flags = append(c.flags, f)
}

// removeFlagByID removes a flag from the command by its ID/"dest".
func (c *Command) removeFlagByID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, f := range c.flags {
		if f.ID == id {
			c.flags = append(c.flags[:i], c.flags[i+1:]...)
			// don't return as there might be multiple.
		}
	}
}

// buildCommand builds the command to be executed. args passed here are any additional
// arguments to be passed to yt-dlp (commonly URLs or similar).
func (c *Command) buildCommand(ctx context.Context, args ...string) *exec.Cmd {
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

	return cmd
}

// runWithResult runs the provided command, collects stdout/stderr, massages the
// result into a Result struct, and returns it (with error wrapping).
func (c *Command) runWithResult(cmd *exec.Cmd) (*Result, error) {
	stdout := &timestampWriter{pipe: "stdout"}
	stderr := &timestampWriter{pipe: "stderr"}

	if slices.Contains(cmd.Args, "--print-json") {
		stdout.checkJSON = true
		stderr.checkJSON = true
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	result := &Result{
		Executable: cmd.Path,
		Args:       cmd.Args[1:],
		ExitCode:   cmd.ProcessState.ExitCode(),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		OutputLogs: stdout.mergeResults(stderr),
	}

	return wrapError(result, err)
}

// Run invokes yt-dlp with the provided arguments (and any flags previously set),
// and returns the results (stdout/stderr, exit code, etc). args should be the
// URLs that would normally be passed in to yt-dlp.
func (c *Command) Run(ctx context.Context, args ...string) (*Result, error) {
	cmd := c.buildCommand(ctx, args...)
	return c.runWithResult(cmd)
}

type Flag struct {
	ID   string   `json:"id"`   // Unique ID to ensure boolean flags are not duplicated.
	Flag string   `json:"flag"` // Actual flag, e.g. "--version".
	Args []string `json:"args"` // Optional args. If nil, it's a boolean flag.
}

func (f *Flag) Clone() *Flag {
	return &Flag{
		ID:   f.ID,
		Flag: f.Flag,
		Args: f.Args,
	}
}

func (f *Flag) Raw() (args []string) {
	args = append(args, f.Flag)
	if f.Args == nil {
		return []string{f.Flag}
	}

	if len(f.Args) > 0 {
		args = append(args, f.Args...)
	}

	return args
}

type Result struct {
	Executable string       `json:"executable"`
	Args       []string     `json:"args"`
	ExitCode   int          `json:"exit_code"`
	Stdout     string       `json:"stdout"`
	Stderr     string       `json:"stderr"`
	OutputLogs []*ResultLog `json:"output_logs"`
}
