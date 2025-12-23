// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	denoResolveCache = atomic.Pointer[ResolvedInstall]{} // Should only be used by [InstallDeno].
	denoInstallLock  sync.Mutex

	denoBinConfigs = map[string]denoBinConfig{
		"darwin_amd64": {
			url:  "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-apple-darwin.zip",
			deno: "deno",
		},
		"darwin_arm64": {
			url:  "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-apple-darwin.zip",
			deno: "deno",
		},
		"linux_amd64": {
			url:  "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip",
			deno: "deno",
		},
		"linux_arm64": {
			url:  "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-unknown-linux-gnu.zip",
			deno: "deno",
		},
		"windows_amd64": {
			url:  "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-pc-windows-msvc.zip",
			deno: "deno.exe",
		},
	}
)

type denoBinConfig struct {
	url  string
	deno string
}

type InstallDenoOptions struct {
	// DisableDownload is a simple toggle to never allow downloading, which would
	// be the same as never calling [InstallDeno] or [MustInstallDeno] in the first place.
	DisableDownload bool

	// DisableSystem is a simple toggle to never allow resolving from the system PATH.
	DisableSystem bool

	// DownloadURL is the exact url to the binary location to download (and store).
	// Leave empty to use GitHub (windows, linux) and evermeet.cx (macos) +
	// auto-detected os/arch.
	DownloadURL string
}

// MustInstallDeno is similar to [InstallDeno], but panics if there is an error.
func MustInstallDeno(ctx context.Context, opts *InstallDenoOptions) {
	_, err := InstallDeno(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// InstallDeno will attempt to download and install deno for the current platform.
// If the binary is already installed or found in the PATH, it will return the resolved
// binary unless [InstallDenoOptions.DisableSystem] is set to true. Note that
// downloading of deno is only supported on a handful of platforms, and so
// it is still recommended to install deno via other means.
func InstallDeno(ctx context.Context, opts *InstallDenoOptions) (*ResolvedInstall, error) {
	denoInstallLock.Lock()
	defer denoInstallLock.Unlock()

	if opts == nil {
		opts = &InstallDenoOptions{}
	}

	if cached := denoResolveCache.Load(); cached != nil {
		return cached, nil
	}

	_, binaries, _ := denoGetDownloadBinary() // don't check error yet.
	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, binaries)
	if err == nil {
		if resolved.Version == "" {
			err = denoGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}

		denoResolveCache.Store(resolved)
		return resolved, nil
	}

	if opts.DisableDownload {
		return nil, errors.New("deno binary not found, and downloading is disabled")
	}

	// Download and install deno.
	resolved, err = downloadAndInstallDeno(ctx, opts)
	if err != nil {
		return nil, err
	}

	denoResolveCache.Store(resolved)
	return resolved, nil
}

func downloadAndInstallDeno(ctx context.Context, opts *InstallDenoOptions) (*ResolvedInstall, error) {
	src, destBinaries, err := denoGetDownloadBinary()
	if err != nil {
		return nil, err
	}

	config, ok := denoBinConfigs[src]
	if !ok {
		return nil, fmt.Errorf("no deno download configuration for %s", src)
	}

	cacheDir, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	downloadURL := opts.DownloadURL
	if downloadURL == "" {
		downloadURL = config.url
	}

	destPath := filepath.Join(cacheDir, destBinaries[0])

	// Download and extract archive.
	err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.deno})
	if err != nil {
		return nil, fmt.Errorf("failed to download and extract deno archive: %w", err)
	}

	return &ResolvedInstall{
		Executable: destPath,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func denoGetDownloadBinary() (src string, dest []string, err error) {
	src = runtime.GOOS + "_" + runtime.GOARCH
	if binary, ok := denoBinConfigs[src]; ok {
		return src, []string{binary.deno}, nil
	}

	var supported []string
	for k := range denoBinConfigs {
		supported = append(supported, k)
	}

	if runtime.GOOS == "windows" {
		dest = []string{"deno.exe"}
	} else {
		dest = []string{"deno"}
	}

	return src, dest, fmt.Errorf(
		"unsupported os/arch combo: %s/%s (supported: %s)",
		runtime.GOOS,
		runtime.GOARCH,
		strings.Join(supported, ", "),
	)
}

var denoVersionRegex = regexp.MustCompile(`^deno ([^ ]+) .*`)

func denoGetVersion(ctx context.Context, r *ResolvedInstall) error {
	var stdout bytes.Buffer

	cmd := exec.Command(r.Executable, "--version") //nolint:gosec
	cmd.Stdout = &stdout
	applySyscall(cmd, false)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run %s to verify version: %w", r.Executable, err)
	}

	version := denoVersionRegex.FindStringSubmatch(stdout.String())
	if len(version) < 2 {
		return fmt.Errorf("unable to parse %s version from output", r.Executable)
	}

	r.Version = version[1]
	debug(ctx, "deno version", "version", r.Version)
	return nil
}
