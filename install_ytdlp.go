// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	//go:embed .github/ytdlp-public.key
	ytdlpPublicKey []byte // From: https://github.com/yt-dlp/yt-dlp/blob/master/public.key

	ytdlpResolveCache   = atomic.Pointer[ResolvedInstall]{} // Should only be used by [Install].
	ytdlpBinConfigCache = atomic.Pointer[string]{}
	ytdlpInstallLock    sync.Mutex

	ytdlpBinConfigs = map[string]struct {
		src  string
		dest []string
	}{
		"darwin_amd64":    {"yt-dlp_macos", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"darwin_arm64":    {"yt-dlp_macos", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_amd64":     {"yt-dlp_linux", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_arm64":     {"yt-dlp_linux_aarch64", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"musllinux_amd64": {"yt-dlp_musllinux", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"musllinux_arm64": {"yt-dlp_musllinux_aarch64", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_armv7l":    {"yt-dlp_linux_armv7l", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_unknown":   {"yt-dlp", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"windows_amd64":   {"yt-dlp.exe", []string{"yt-dlp-" + Version + ".exe", "yt-dlp.exe"}},
	}
)

// ytdlpGetDownloadBinary returns the source and destination binary names for the
// current runtime. If the current runtime is not supported, an error is
// returned. dest will always be returned (it will be an assumption).
func ytdlpGetDownloadBinary() (src string, dest []string, err error) {
	if cached := ytdlpBinConfigCache.Load(); cached != nil {
		if binary, ok := ytdlpBinConfigs[*cached]; ok {
			return binary.src, binary.dest, nil
		}
	}

	src = runtime.GOOS + "_" + runtime.GOARCH
	if runtime.GOOS == "linux" && systemHasMusl() {
		src = "musl" + src
	}

	if binary, ok := ytdlpBinConfigs[src]; ok {
		ytdlpBinConfigCache.Store(&src)
		return binary.src, binary.dest, nil
	}

	if runtime.GOOS == "linux" {
		src = "linux_unknown"
		ytdlpBinConfigCache.Store(&src)
		return ytdlpBinConfigs[src].src, ytdlpBinConfigs[src].dest, nil
	}

	var supported []string
	for k := range ytdlpBinConfigs {
		supported = append(supported, k)
	}

	if runtime.GOOS == "windows" {
		dest = []string{"yt-dlp.exe"}
	} else {
		dest = []string{"yt-dlp"}
	}

	return "", dest, fmt.Errorf(
		"unsupported os/arch combo: %s/%s (supported: %s)",
		runtime.GOOS,
		runtime.GOARCH,
		strings.Join(supported, ", "),
	)
}

// InstallOptions are configuration options for installing yt-dlp dynamically (when
// it's not already installed).
type InstallOptions struct {
	// DisableDownload is a simple toggle to never allow downloading, which would
	// be the same as never calling [Install] or [MustInstall] in the first place.
	DisableDownload bool

	// DisableChecksum disables checksum verification when downloading.
	DisableChecksum bool

	// DisableSystem is a simple toggle to never allow resolving from the system PATH.
	DisableSystem bool

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

func ytdlpGithubReleaseAsset(name string) string {
	return fmt.Sprintf("https://github.com/yt-dlp/yt-dlp/releases/download/%s/%s", Version, name)
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

	if r := ytdlpResolveCache.Load(); r != nil {
		return r, nil
	}

	// Ensure only one install invocation is running at a time.
	ytdlpInstallLock.Lock()
	defer ytdlpInstallLock.Unlock()

	_, binaries, _ := ytdlpGetDownloadBinary() // don't check error yet.
	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, binaries)
	if err == nil {
		if resolved.Version == "" {
			err = ytdlpGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}

		if opts.AllowVersionMismatch {
			ytdlpResolveCache.Store(resolved)
			return resolved, nil
		}

		if resolved.Version == Version {
			ytdlpResolveCache.Store(resolved)
			return resolved, nil
		}

		// If we're not allowed to download, and the version doesn't match, return
		// an error.
		if opts.DisableDownload {
			return nil, fmt.Errorf("yt-dlp version mismatch: expected %s, got %s", Version, resolved.Version)
		}
	}

	if opts.DisableDownload {
		return nil, errors.New("yt-dlp executable not found, and downloading is disabled")
	}

	src, dest, err := ytdlpGetDownloadBinary()
	if err != nil {
		return nil, err
	}

	downloadURL := opts.DownloadURL

	if downloadURL == "" {
		downloadURL = ytdlpGithubReleaseAsset(src)
	}

	// Prepare cache directory.
	dir, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	_, err = downloadFile(ctx, downloadURL, dir, filepath.Join(dir, dest[0]+".tmp"), 0o700) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	if !opts.DisableChecksum {
		_, err = downloadFile(
			ctx,
			ytdlpGithubReleaseAsset("SHA2-256SUMS"),
			dir,
			filepath.Join(dir, "SHA2-256SUMS-"+Version),
			0o700,
		) //nolint:gomnd
		if err != nil {
			return nil, err
		}

		_, err = downloadFile(
			ctx,
			ytdlpGithubReleaseAsset("SHA2-256SUMS.sig"),
			dir,
			filepath.Join(dir, "SHA2-256SUMS-"+Version+".sig"),
			0o700,
		) //nolint:gomnd
		if err != nil {
			return nil, err
		}

		err = verifyFileChecksum(
			ctx,
			filepath.Join(dir, "SHA2-256SUMS-"+Version),
			filepath.Join(dir, "SHA2-256SUMS-"+Version+".sig"),
			filepath.Join(dir, dest[0]+".tmp"),
			src,
		)
		if err != nil {
			return nil, err
		}
	}

	// Rename the file to the correct name.
	err = os.Rename(filepath.Join(dir, dest[0]+".tmp"), filepath.Join(dir, dest[0]))
	if err != nil {
		return nil, fmt.Errorf("unable to rename yt-dlp executable: %w", err)
	}

	// re-resolve now that we've downloaded the binary, and validated things.
	resolved, err = resolveExecutable(ctx, true, opts.DisableSystem, binaries)
	if err != nil {
		return nil, err
	}

	if resolved.Version == "" {
		err = ytdlpGetVersion(ctx, resolved)
		if err != nil {
			return nil, err
		}
	}

	ytdlpResolveCache.Store(resolved)
	return resolved, nil
}

// MustInstall is the same as [Install], but will panic if an error occurs (essentially
// ensuring yt-dlp is installed, before continuing), and doesn't return any results.
func MustInstall(ctx context.Context, opts *InstallOptions) {
	_, err := Install(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// ytdlpGetVersion sets the version of the resolved executable.
func ytdlpGetVersion(ctx context.Context, r *ResolvedInstall) error {
	var stdout bytes.Buffer

	cmd := exec.Command(r.Executable, "--version") //nolint:gosec
	cmd.Stdout = &stdout
	applySyscall(cmd, false)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run yt-dlp to verify version: %w", err)
	}

	r.Version = strings.TrimSpace(stdout.String())
	debug(ctx, "yt-dlp version", "version", r.Version)
	return nil
}

// systemHasMusl reports whether the system uses musl libc.
func systemHasMusl() bool {
	cmd := exec.Command("ldd", "--version")
	// Ignore error because ldd often returns exit status 1 even with valid output
	output, _ := cmd.CombinedOutput()

	if len(output) > 0 {
		// ldd outputs libc info in first line: "musl libc..."
		firstLine := strings.Split(string(output), "\n")[0]
		return strings.Contains(strings.ToLower(firstLine), "musl")
	}

	// Fallback: check common musl lib paths
	muslPaths := []string{
		"/lib/ld-musl-x86_64.so.1",
		"/lib/ld-musl-aarch64.so.1",
		"/lib/ld-musl-armhf.so.1",
	}

	for _, path := range muslPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
