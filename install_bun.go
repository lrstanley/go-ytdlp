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
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	bunResolveCache = atomic.Pointer[ResolvedInstall]{} // Should only be used by [InstallBun].
	bunInstallLock  sync.Mutex

	bunBinConfigs = map[string]bunBinConfig{
		"darwin_amd64": {
			url: "https://github.com/oven-sh/bun/releases/latest/download/bun-darwin-x64.zip",
			bun: "bun",
		},
		"darwin_arm64": {
			url: "https://github.com/oven-sh/bun/releases/latest/download/bun-darwin-aarch64.zip",
			bun: "bun",
		},
		"linux_amd64": {
			url: "https://github.com/oven-sh/bun/releases/latest/download/bun-linux-x64.zip",
			bun: "bun",
		},
		"linux_arm64": {
			url: "https://github.com/oven-sh/bun/releases/latest/download/bun-linux-aarch64.zip",
			bun: "bun",
		},
		"windows_amd64": {
			url: "https://github.com/oven-sh/bun/releases/latest/download/bun-windows-x64.zip",
			bun: "bun.exe",
		},
	}
)

type bunBinConfig struct {
	url string
	bun string
}

type InstallBunOptions struct {
	// DisableDownload is a simple toggle to never allow downloading, which would
	// be the same as never calling [InstallBun] or [MustInstallBun] in the first place.
	DisableDownload bool

	// DisableSystem is a simple toggle to never allow resolving from the system PATH.
	DisableSystem bool

	// DownloadURL is the exact url to the binary location to download (and store).
	// Leave empty to use GitHub (windows, linux) and evermeet.cx (macos) +
	// auto-detected os/arch.
	DownloadURL string
}

// MustInstallBun is similar to [InstallBun], but panics if there is an error.
func MustInstallBun(ctx context.Context, opts *InstallBunOptions) {
	_, err := InstallBun(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// InstallBun will attempt to download and install bun for the current platform.
// If the binary is already installed or found in the PATH, it will return the resolved
// binary unless [InstallBunOptions.DisableSystem] is set to true. Note that
// downloading of bun is only supported on a handful of platforms, and so
// it is still recommended to install bun via other means.
func InstallBun(ctx context.Context, opts *InstallBunOptions) (*ResolvedInstall, error) {
	bunInstallLock.Lock()
	defer bunInstallLock.Unlock()

	if opts == nil {
		opts = &InstallBunOptions{}
	}

	if cached := bunResolveCache.Load(); cached != nil {
		return cached, nil
	}

	_, binaries, _ := bunGetDownloadBinary() // don't check error yet.
	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, binaries)
	if err == nil {
		if resolved.Version == "" {
			err = bunGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}

		bunResolveCache.Store(resolved)
		return resolved, nil
	}

	if opts.DisableDownload {
		return nil, errors.New("bun binary not found, and downloading is disabled")
	}

	// Download and install bun.
	resolved, err = downloadAndInstallBun(ctx, opts)
	if err != nil {
		return nil, err
	}

	bunResolveCache.Store(resolved)
	return resolved, nil
}

func downloadAndInstallBun(ctx context.Context, opts *InstallBunOptions) (*ResolvedInstall, error) {
	src, destBinaries, err := bunGetDownloadBinary()
	if err != nil {
		return nil, err
	}

	config, ok := bunBinConfigs[src]
	if !ok {
		return nil, fmt.Errorf("no bun download configuration for %s", src)
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
	err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.bun})
	if err != nil {
		return nil, fmt.Errorf("failed to download and extract bun archive: %w", err)
	}

	return &ResolvedInstall{
		Executable: destPath,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func bunGetDownloadBinary() (src string, dest []string, err error) {
	src = runtime.GOOS + "_" + runtime.GOARCH
	if binary, ok := bunBinConfigs[src]; ok {
		return src, []string{binary.bun}, nil
	}

	var supported []string
	for k := range bunBinConfigs {
		supported = append(supported, k)
	}

	if runtime.GOOS == "windows" {
		dest = []string{"bun.exe"}
	} else {
		dest = []string{"bun"}
	}

	return src, dest, fmt.Errorf(
		"unsupported os/arch combo: %s/%s (supported: %s)",
		runtime.GOOS,
		runtime.GOARCH,
		strings.Join(supported, ", "),
	)
}

func bunGetVersion(ctx context.Context, r *ResolvedInstall) error {
	var stdout bytes.Buffer

	cmd := exec.Command(r.Executable, "--version") //nolint:gosec
	cmd.Stdout = &stdout
	applySyscall(cmd, false)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run %s to verify version: %w", r.Executable, err)
	}

	r.Version = strings.TrimSpace(stdout.String())
	debug(ctx, "bun version", "version", r.Version)
	return nil
}
