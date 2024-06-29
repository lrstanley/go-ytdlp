// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
)

// New is the recommended way to return a new yt-dlp command builder. Once all
// flags are set, you can call [Run] to invoke yt-dlp with the necessary args, or
// the independent execution method (e.g. [Version]).
func New() *Command {
	cmd := &Command{
		env: make(map[string]string),
	}

	return cmd
}

type Command struct {
	mu           sync.RWMutex
	executable   string
	directory    string
	env          map[string]string
	flags        []*Flag
	progressFunc DownloadProgressFunc
}

// Clone returns a copy of the command, with all flags, env vars, executable, and
// working directory copied over.
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

func (c *Command) hasJSONFlag() bool {
	pf := c.getFlagsByID("forceprint")

	return (len(pf) > 0 && len(pf[0].Args) > 0 && pf[0].Args[0] == "%()s") ||
		c.getFlagsByID("print_json") != nil ||
		c.getFlagsByID("dumpjson") != nil
}

// SetProgressFn sets the progress function to be called on each progress event.
func (c *Command) SetProgressFn(progressFunc DownloadProgressFunc) *Command {
	c.progressFunc = progressFunc
	return c
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
	var err error

	c.mu.RLock()
	name = c.executable

	if name == "" {
		var r *ResolvedInstall
		r, err = resolveExecutable(true, false)
		if err == nil {
			name = r.Executable
		}
	}

	cmd := exec.CommandContext(ctx, name, cmdArgs...)

	if err != nil {
		cmd.Err = err // Hijack the existing command to return the error from resolveExecutable.
	}

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
	if cmd.Err != nil {
		return wrapError(nil, cmd.Err)
	}

	stdout := &timestampWriter{
		pipe:         "stdout",
		downloads:    make(map[string]*DownloadProgress),
		progressFunc: c.progressFunc,
	}
	stderr := &timestampWriter{pipe: "stderr"}

	if c.hasJSONFlag() {
		stdout.checkJSON = true
		stderr.checkJSON = true
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	c.applySyscall(cmd)
	err := cmd.Run()

	result := &Result{
		Executable: cmd.Path,
		Args:       cmd.Args[1:],
		ExitCode:   cmd.ProcessState.ExitCode(),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		OutputLogs: stdout.mergeResults(stderr),
		Downloads:  stdout.GetDownloads(),
	}

	return wrapError(result, err)
}

// Run invokes yt-dlp with the provided arguments (and any flags previously set),
// and returns the results (stdout/stderr, exit code, etc). args should be the
// URLs that would normally be passed in to yt-dlp.
func (c *Command) Run(ctx context.Context, args ...string) (*Result, error) {
	if c.progressFunc != nil {
		c.Progress()

		progressTempl, err := GetDownloadProgressTemplate()
		if err != nil {
			return nil, fmt.Errorf("failed to make download progress template: %w", err)
		}
		c.ProgressTemplate(progressTempl)

		// Enables yt-dlp to print progress updates in a new line,
		// instead of using carriage return to overwrite the previous line. This ensures
		// the progress updates can be processed line by line and we won't miss any updates.
		c.Newline()

		// preProcessTempl, err := GetProgressPreProcessTemplate()
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to make pre-process template: %w", err)
		// }
		// c.Print(preProcessTempl)

		beforeDownloadTempl, err := GetProgressBeforeDownloadTemplate()
		if err != nil {
			return nil, fmt.Errorf("failed to make before-download template: %w", err)
		}
		c.Print(beforeDownloadTempl)

		postProcessTempl, err := GetProgressPostProcessTemplate()
		if err != nil {
			return nil, fmt.Errorf("failed to make post-process template: %w", err)
		}
		c.Print(postProcessTempl)

		videoDownloadedTempl, err := GetProgressVideoDownloadedTemplate()
		if err != nil {
			return nil, fmt.Errorf("failed to make after-video template: %w", err)
		}
		c.Print(videoDownloadedTempl)

		// playListDownloadedTempl, err := GetProgressPlaylistDownloadedTemplate()
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to make playlist template: %w", err)
		// }
		// c.Print(playListDownloadedTempl)
	}

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
