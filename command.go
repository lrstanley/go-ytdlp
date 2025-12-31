// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// New is the recommended way to return a new yt-dlp command builder. Once all
// flags are set, you can call [Run] to invoke yt-dlp with the necessary args, or
// the independent execution method (e.g. [Version]).
func New() *Command {
	cmd := &Command{
		env:           make(map[string]string),
		flagConfig:    &FlagConfig{},
		cancelMaxWait: 1 * time.Second,
	}
	return cmd
}

type Command struct {
	mu                   sync.RWMutex
	executable           string
	directory            string
	env                  map[string]string
	flagConfig           *FlagConfig
	separateProcessGroup bool
	cancelMaxWait        time.Duration
	disableEnvVarInherit bool

	progress *progressHandler
}

// Clone returns a copy of the command, with all flags, env vars, executable,
// working directory, etc copied over.
func (c *Command) Clone() *Command {
	c.mu.RLock()
	cc := &Command{
		executable:           c.executable,
		directory:            c.directory,
		env:                  make(map[string]string, len(c.env)),
		flagConfig:           c.flagConfig.Clone(),
		separateProcessGroup: c.separateProcessGroup,
		progress:             c.progress,
	}
	maps.Copy(cc.env, c.env)
	c.mu.RUnlock()
	return cc
}

// GetFlagConfig returns a copy of the flag config.
func (c *Command) GetFlagConfig() *FlagConfig {
	c.mu.RLock()
	cc := c.flagConfig.Clone()
	c.mu.RUnlock()
	return cc
}

// SetFlagConfig sets the flag config for the command, overriding ALL previously
// set flags. If nil is provided, a new empty flag config will be used.
func (c *Command) SetFlagConfig(flagConfig *FlagConfig) *Command {
	if flagConfig == nil {
		flagConfig = &FlagConfig{}
	}
	c.mu.Lock()
	c.flagConfig = flagConfig.Clone()
	c.mu.Unlock()
	return c
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
// be removed. If the key is "PATH", it will be merged with the parent process's PATH,
// (where explicit values provided here will take precedence). See also [SetEnvVarInherit].
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

// SetSeparateProcessGroup sets whether the command should be run in a separate
// process group. This is useful to avoid propagating signals from the app process.
// NOTE: This is only supported on Windows and Unix-like systems.
func (c *Command) SetSeparateProcessGroup(value bool) *Command {
	c.mu.Lock()
	c.separateProcessGroup = value
	c.mu.Unlock()

	return c
}

// SetCancelMaxWait sets the maximum wait time before the command is killed,
// after the context is cancelled. Defaults to 1 second.
func (c *Command) SetCancelMaxWait(value time.Duration) *Command {
	c.mu.Lock()
	c.cancelMaxWait = value
	c.mu.Unlock()
	return c
}

// SetEnvVarInherit sets whether the command should inherit environment variables
// from the parent process. If enabled, the command will inherit the parent process's
// environment variables, and any environment variables set with [SetEnvVar] will be
// merged with the parent process's environment variables. If disabled, the command
// will only use the environment variables set with [SetEnvVar] (with the exception
// of PATH). Explicitly set env vars with [SetEnvVar] will always take precedence
// over the parent process's environment variables.
func (c *Command) SetEnvVarInherit(enabled bool) *Command {
	c.mu.Lock()
	c.disableEnvVarInherit = !enabled
	c.mu.Unlock()
	return c
}

func (c *Command) hasJSONFlag() bool {
	if slices.Contains(c.flagConfig.VerbositySimulation.Print, "%(j)") && len(c.flagConfig.VerbositySimulation.Print) == 1 {
		return true
	}
	if v := c.flagConfig.VerbositySimulation.PrintJSON; v != nil && *v {
		return true
	}
	if v := c.flagConfig.VerbositySimulation.DumpJSON; v != nil && *v {
		return true
	}
	return false
}

// toMap converts a slice of environment variables to a map. Handles Windows
// environment variables that start with '='.
func toMap(env []string) map[string]string {
	r := map[string]string{}
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)

		if runtime.GOOS == "windows" {
			// On Windows, env vars can start with "=".
			prefix := false
			if len(e) > 0 && e[0] == '=' {
				e = e[1:]
				prefix = true
			}
			p = strings.SplitN(e, "=", 2)
			if prefix {
				p[0] = "=" + p[0]
			}
		}

		if len(p) == 2 {
			r[p[0]] = p[1]
		}
	}
	return r
}

// BuildCommand builds the command to be executed. args passed here are any additional
// arguments to be passed to yt-dlp (commonly URLs or similar). This should not be used
// directly unless you want to reference the arguments passed to yt-dlp.
func (c *Command) BuildCommand(ctx context.Context, args ...string) *exec.Cmd {
	var cmdArgs []string

	// If bun is in resolve cache and user didn't set any runtime settings, enable bun
	if bunResolveCache.Load() != nil &&
		len(c.flagConfig.General.JsRuntimes) == 0 &&
		c.flagConfig.General.NoJsRuntimes == nil {

		debug(ctx, "automatically enabling js runtime bun")
		c.flagConfig.General.JsRuntimes = []string{"bun"}
	}

	for _, f := range c.flagConfig.ToFlags() {
		cmdArgs = append(cmdArgs, f.Raw()...)
	}

	cmdArgs = append(cmdArgs, args...) // URLs or similar.

	var name string
	var err error

	c.mu.RLock()
	name = c.executable

	if name == "" {
		r := ytdlpResolveCache.Load()
		if r == nil {
			_, binaries, _ := ytdlpGetDownloadBinary() // don't check error yet.
			r, err = resolveExecutable(ctx, false, false, binaries)
			if err == nil {
				name = r.Executable
			}
		} else {
			name = r.Executable
		}
	}

	env := map[string]string{}
	if !c.disableEnvVarInherit {
		env = toMap(os.Environ())
	}

	// Merge in the command's environment variables, accounting for things like
	// PATH, which should be merged with the previously provided PATH.
	for k, v := range c.env {
		switch k {
		case "PATH":
			if env["PATH"] != "" {
				paths := filepath.SplitList(env["PATH"])
				cpaths := filepath.SplitList(v)
				// Append parent process paths to the end of our custom provided
				// paths, only if they are not already in the PATH.
				for _, p := range paths {
					if !slices.Contains(cpaths, p) {
						cpaths = append([]string{p}, cpaths...)
					}
				}
				env["PATH"] = strings.Join(cpaths, string(filepath.ListSeparator))
			} else {
				env["PATH"] = v
			}
		default:
			env[k] = v
		}
	}

	cmd := exec.CommandContext(ctx, name, cmdArgs...)

	// Ensure all children (e.g. ffmpeg) are killed after the command is killed.
	cmd.WaitDelay = c.cancelMaxWait

	// Add cache directory to $PATH, which would cover ffmpeg, ffprobe, etc.
	cacheDir, err := GetCacheDir()
	if err == nil {
		env["PATH"] = strings.Join(append([]string{cacheDir}, filepath.SplitList(env["PATH"])...), string(filepath.ListSeparator))
	}

	if err != nil {
		cmd.Err = err // Hijack the existing command to return the error from resolveExecutable.
	}

	if c.directory != "" {
		cmd.Dir = c.directory
	}

	cmd.Env = make([]string, 0, len(env))
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	c.mu.RUnlock()

	return cmd
}

// runWithResult runs the provided command, collects stdout/stderr, massages the
// result into a Result struct, and returns it (with error wrapping).
func (c *Command) runWithResult(ctx context.Context, cmd *exec.Cmd) (*Result, error) {
	if cmd.Err != nil {
		return wrapError(nil, cmd.Err)
	}

	stdout := &timestampWriter{pipe: "stdout", progress: c.progress}
	stderr := &timestampWriter{pipe: "stderr"}

	if c.hasJSONFlag() {
		stdout.checkJSON = true
		stderr.checkJSON = true
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	applySyscall(cmd, c.separateProcessGroup)

	debug(
		ctx, "running command",
		"path", cmd.Path,
		"args", cmd.Args,
		"env", cmd.Env,
		"dir", cmd.Dir,
	)

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
	if err := c.flagConfig.Validate(); err != nil {
		return nil, err
	}

	cmd := c.BuildCommand(ctx, args...)
	return c.runWithResult(ctx, cmd)
}

type Flag struct {
	ID             string `json:"id"`              // Unique ID to ensure boolean flags are not duplicated.
	Flag           string `json:"flag"`            // Actual flag, e.g. "--version".
	AllowsMultiple bool   `json:"allows_multiple"` // If the flag allows multiple values.
	Args           []any  `json:"args"`            // Optional args. If nil, it's a boolean flag.
}

func (f *Flag) Raw() (args []string) {
	args = append(args, f.Flag)
	if f.Args == nil {
		return args
	}

	for _, arg := range f.Args {
		if arg == nil {
			continue
		}

		switch arg := arg.(type) {
		case string:
			args = append(args, arg)
		case int:
			args = append(args, strconv.Itoa(arg))
		case int64:
			args = append(args, strconv.FormatInt(arg, 10))
		case float64:
			args = append(args, strconv.FormatFloat(arg, 'g', -1, 64))
		case bool:
			args = append(args, strconv.FormatBool(arg))
		default:
			panic(fmt.Sprintf("unsupported arg type for flag: %T", arg))
		}
	}

	return args
}

type Flags []*Flag

func (f Flags) FindByID(id string) (flags Flags) {
	for _, flag := range f {
		if flag.ID == id {
			flags = append(flags, flag)
		}
	}
	return flags
}

func (f Flags) Duplicates() (duplicates Flags) {
	seen := make(map[string]Flags)
	for _, flag := range f {
		if flag.AllowsMultiple {
			continue
		}
		seen[flag.ID] = append(seen[flag.ID], flag)
	}
	for _, flags := range seen {
		if len(flags) > 1 {
			duplicates = append(duplicates, flags...)
		}
	}
	return duplicates
}
