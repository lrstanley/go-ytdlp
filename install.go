// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	xdgCacheDir     = "go-ytdlp"       // Cache directory that will be appended to the XDG cache directory.
	downloadTimeout = 30 * time.Second // HTTP timeout for downloading the yt-dlp binary.
)

var (
	resolveCache = atomic.Pointer[ResolvedInstall]{} // Should only be used by [Install].
	installLock  sync.Mutex

	binConfigs = map[string]struct {
		src  string
		dest []string
	}{
		"darwin_amd64":  {"yt-dlp_macos", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"darwin_arm64":  {"yt-dlp_macos", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_amd64":   {"yt-dlp_linux", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_arm64":   {"yt-dlp_linux_armv7l", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_armv7l":  {"yt-dlp_linux_armv7l", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_unknown": {"yt-dlp", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"windows_amd64": {"yt-dlp.exe", []string{"yt-dlp-" + Version + ".exe", "yt-dlp.exe"}},
	}
)

// getDownloadBinary returns the source and destination binary names for the
// current runtime. If the current runtime is not supported, an error is
// returned. dest will always be returned (it will be an assumption).
func getDownloadBinary() (src string, dest []string, err error) {
	if binary, ok := binConfigs[runtime.GOOS+"_"+runtime.GOARCH]; ok {
		return binary.src, binary.dest, nil
	}

	if runtime.GOOS == "linux" {
		return binConfigs["linux_unknown"].src, binConfigs["linux_unknown"].dest, nil
	}

	var supported []string
	for k := range binConfigs {
		supported = append(supported, k)
	}

	if runtime.GOOS == "windows" {
		dest = []string{"yt-dlp.exe"}
	} else {
		dest = []string{"yt-dlp"}
	}

	return "", dest, fmt.Errorf("unsupported os/arch combo: %s/%s (supported: %s)", runtime.GOOS, runtime.GOARCH, strings.Join(supported, ", "))
}

// InstallOptions are configuration options for installing yt-dlp dynamically (when
// it's not already installed).
type InstallOptions struct {
	// DisableDownload is a simple toggle to never allow downloading, which would
	// be the same as never calling [Install] or [MustInstall] in the first place.
	DisableDownload bool

	// AllowVersionMismatch allows mismatched versions to be used and installed.
	// This will only be used when the yt-dlp executable is resolved outside of
	// go-ytdlp's cache.
	//
	// AllowVersionMismatch is ignored if DisableDownload is true.
	AllowVersionMismatch bool

	// DownloadURL is the exact url to the binary location to download (and store).
	// Leave empty to use GitHub + auto-detected os/arch.
	DownloadURL string
}

// Install will check to see if yt-dlp is installed (if it's the right version),
// and if not, will download it from GitHub. If yt-dlp is already installed, it will
// check to see if the version matches (unless disabled with [AllowVersionMismatch]),
// and if not, will download the same version that go-ytdlp (the version you are using)
// was built with.
//
// Note: If [Install] is not called, go-ytdlp WILL NOT DOWNLOAD yt-dlp. Only use
// this function if you want to ensure yt-dlp is installed, and are ok with it being
// downloaded.
func Install(ctx context.Context, opts *InstallOptions) (*ResolvedInstall, error) {
	if opts == nil {
		opts = &InstallOptions{}
	}

	if r := resolveCache.Load(); r != nil {
		return r, nil
	}

	// Ensure only one install invocation is running at a time.
	installLock.Lock()
	defer installLock.Unlock()

	resolved, err := resolveExecutable(false, false)
	if err == nil {
		if opts.AllowVersionMismatch {
			resolveCache.Store(resolved)
			return resolved, nil
		}

		if resolved.Version == Version {
			resolveCache.Store(resolved)
			return resolved, nil
		}

		// If we're not allowed to download, and the version doesn't match, return
		// an error.
		if opts.DisableDownload {
			return nil, fmt.Errorf("yt-dlp version mismatch: expected %s, got %s", Version, resolved.Version)
		}
	}

	if opts.DisableDownload {
		return nil, fmt.Errorf("yt-dlp executable not found, and downloading is disabled")
	}

	src, dest, err := getDownloadBinary()
	if err != nil {
		return nil, err
	}

	downloadURL := opts.DownloadURL

	if downloadURL == "" {
		downloadURL = fmt.Sprintf("https://github.com/yt-dlp/yt-dlp/releases/download/%s/%s", Version, src)
	}

	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get user cache dir: %w", err)
	}
	dir := filepath.Join(baseCacheDir, xdgCacheDir)

	err = os.MkdirAll(dir, 0o750)
	if err != nil {
		return nil, fmt.Errorf("unable to create yt-dlp executable cache directory: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(dir, dest[0]), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o750)
	if err != nil {
		return nil, fmt.Errorf("unable to create yt-dlp executable cache file: %w", err)
	}
	defer f.Close()

	// Download the binary.
	client := &http.Client{Timeout: downloadTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to download yt-dlp executable: request creation: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("github.com/lrstanley/go-ytdlp; version/%s", Version))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to download yt-dlp executable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to download yt-dlp executable: bad status: %s", resp.Status)
	}

	_, err = f.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to download yt-dlp executable: streaming data: %w", err)
	}

	err = f.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to download yt-dlp executable: closing file: %w", err)
	}

	resolved, err = resolveExecutable(false, true)
	if err != nil {
		return nil, err
	}

	resolveCache.Store(resolved)
	return resolved, nil
}

// MustInstall is the same as [Install], but will panic if an error occurs (essentially
// ensuring yt-dlp is installed, before continuing).
func MustInstall(ctx context.Context, opts *InstallOptions) *ResolvedInstall {
	r, err := Install(ctx, opts)
	if err != nil {
		panic(err)
	}
	return r
}

// resolveExecutable will attempt to resolve the yt-dlp executable, either from
// the go-ytdlp cache (first), or from the PATH (second). If it's not found, an
// error is returned.
func resolveExecutable(fromCache, calleeIsDownloader bool) (r *ResolvedInstall, err error) {
	if fromCache {
		r = resolveCache.Load()
		if r != nil {
			return r, nil
		}
	}

	_, dest, _ := getDownloadBinary() // don't check error yet.

	var stat os.FileInfo
	var bin, baseCacheDir string

	baseCacheDir, err = os.UserCacheDir()
	if err == nil {
		// Check out cache dirs first.
		for _, d := range dest {
			bin = filepath.Join(baseCacheDir, xdgCacheDir, d)

			stat, err = os.Stat(bin)
			if err == nil {
				if !stat.IsDir() && (stat.Mode().Perm()&0o100 != 0 || stat.Mode().Perm()&0o010 != 0 || stat.Mode().Perm()&0o001 != 0) {
					r = &ResolvedInstall{
						Executable: bin,
						FromCache:  true,
						Downloaded: calleeIsDownloader,
					}
					if calleeIsDownloader {
						r.Version = Version
					} else {
						err = r.getVersion()
						if err != nil {
							return nil, err
						}
					}
					return r, nil
				}
			}
		}
	}

	// Check PATH for the binary.
	for _, d := range dest {
		bin, err = exec.LookPath(d)
		if err == nil {
			r = &ResolvedInstall{
				Executable: bin,
				FromCache:  false,
				Downloaded: false,
			}

			err = r.getVersion()
			if err != nil {
				return nil, err
			}

			return r, nil
		}
	}

	// Will pick the last error, which is likely without the version suffix, what we want.
	return nil, fmt.Errorf("unable to resolve yt-dlp executable: %w", err)
}

// ResolvedInstall is the found yt-dlp executable.
type ResolvedInstall struct {
	Executable string // Path to the yt-dlp executable.
	Version    string // Version of yt-dlp that was resolved. If [InstallOptions.AllowVersionMismatch] is specified, this will be empty.
	FromCache  bool   // Whether the executable was resolved from the cache.
	Downloaded bool   // Whether the executable was downloaded during this invocation.
}

// getVersion returns true if the resolved version of yt-dlp matches the version
// that go-ytdlp was built with.
func (r *ResolvedInstall) getVersion() error {
	var stdout bytes.Buffer

	cmd := exec.Command(r.Executable, "--version") //nolint:gosec
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run yt-dlp to verify version: %w", err)
	}

	r.Version = strings.TrimSpace(stdout.String())

	return nil
}
