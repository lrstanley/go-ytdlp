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
	ffmpegResolveCache  = atomic.Pointer[ResolvedInstall]{} // Should only be used by [InstallFFmpeg].
	ffmpegInstallLock   sync.Mutex
	ffprobeResolveCache = atomic.Pointer[ResolvedInstall]{} // Should only be used by [InstallFFprobe].
	ffprobeInstallLock  sync.Mutex

	ffmpegBinConfigs = map[string]ffmpegBinConfig{
		"darwin_amd64": {
			ffmpegURL:  "https://evermeet.cx/ffmpeg/getrelease/ffmpeg",
			ffprobeURL: "https://evermeet.cx/ffmpeg/getrelease/ffprobe",
			ffmpeg:     "ffmpeg",
			ffprobe:    "ffprobe",
			isArchive:  false,
		},
		"linux_amd64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz",
			ffmpeg:    "ffmpeg",
			ffprobe:   "ffprobe",
			isArchive: true,
		},
		"linux_arm64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz",
			ffmpeg:    "ffmpeg",
			ffprobe:   "ffprobe",
			isArchive: true,
		},
		"windows_amd64": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
			ffmpeg:    "ffmpeg.exe",
			ffprobe:   "ffprobe.exe",
			isArchive: true,
		},
		"windows_arm": {
			ffmpegURL: "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-winarm64-gpl.zip",
			ffmpeg:    "ffmpeg.exe",
			ffprobe:   "ffprobe.exe",
			isArchive: true,
		},
	}
)

type ffmpegBinConfig struct {
	ffmpegURL  string
	ffprobeURL string
	ffmpeg     string
	ffprobe    string
	isArchive  bool // true if the download is an archive containing both binaries
}

type InstallFFmpegOptions struct {
	// DisableDownload is a simple toggle to never allow downloading, which would
	// be the same as never calling [InstallFFmpeg] or [MustInstallFFmpeg] in the first place.
	DisableDownload bool

	// DisableSystem is a simple toggle to never allow resolving from the system PATH.
	DisableSystem bool

	// DownloadURL is the exact url to the binary location to download (and store).
	// Leave empty to use GitHub (windows, linux) and evermeet.cx (macos) +
	// auto-detected os/arch.
	DownloadURL string
}

// MustInstallFFmpeg is similar to [InstallFFmpeg], but panics if there is an error.
func MustInstallFFmpeg(ctx context.Context, opts *InstallFFmpegOptions) {
	_, err := InstallFFmpeg(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// InstallFFmpeg will attempt to download and install FFmpeg for the current platform.
// If the binary is already installed or found in the PATH, it will return the resolved
// binary unless [InstallFFmpegOptions.DisableSystem] is set to true. Note that
// downloading of ffmpeg and ffprobe is only supported on a handful of platforms, and so
// it is still recommended to install ffmpeg/ffprobe via other means.
func InstallFFmpeg(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	ffmpegInstallLock.Lock()
	defer ffmpegInstallLock.Unlock()

	if opts == nil {
		opts = &InstallFFmpegOptions{}
	}

	if cached := ffmpegResolveCache.Load(); cached != nil {
		return cached, nil
	}

	_, binaries, _ := ffmpegGetDownloadBinary() // don't check error yet.
	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, binaries)
	if err == nil {
		if resolved.Version == "" {
			err = ffGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}

		ffmpegResolveCache.Store(resolved)
		return resolved, nil
	}

	if opts.DisableDownload {
		return nil, errors.New("ffmpeg binary not found, and downloading is disabled")
	}

	// Download and install FFmpeg (and FFprobe if archive).
	resolved, err = downloadAndInstallFFmpeg(ctx, opts)
	if err != nil {
		return nil, err
	}

	ffmpegResolveCache.Store(resolved)
	return resolved, nil
}

// MustInstallFFprobe is similar to [InstallFFprobe], but panics if there is an error.
func MustInstallFFprobe(ctx context.Context, opts *InstallFFmpegOptions) {
	_, err := InstallFFprobe(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// InstallFFprobe will attempt to download and install FFprobe for the current platform.
// If the binary is already installed or found in the PATH, it will return the resolved
// binary unless [InstallFFmpegOptions.DisableSystem] is set to true. Note that
// downloading of ffmpeg and ffprobe is only supported on a handful of platforms, and so
// it is still recommended to install ffmpeg/ffprobe via other means.
func InstallFFprobe(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	ffprobeInstallLock.Lock()
	defer ffprobeInstallLock.Unlock()

	if opts == nil {
		opts = &InstallFFmpegOptions{}
	}

	if cached := ffprobeResolveCache.Load(); cached != nil {
		return cached, nil
	}

	_, binaries, _ := ffprobeGetDownloadBinary() // don't check error yet.
	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, binaries)
	if err == nil {
		if resolved.Version == "" {
			err = ffGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}

		ffprobeResolveCache.Store(resolved)
		return resolved, nil
	}

	if opts.DisableDownload {
		return nil, errors.New("ffprobe binary not found, and downloading is disabled")
	}

	// Download and install FFprobe (and FFmpeg if archive).
	resolved, err = downloadAndInstallFFprobe(ctx, opts)
	if err != nil {
		return nil, err
	}

	ffprobeResolveCache.Store(resolved)
	return resolved, nil
}

func downloadAndInstallFFmpeg(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	src, destBinaries, err := ffmpegGetDownloadBinary()
	if err != nil {
		return nil, err
	}

	config, ok := ffmpegBinConfigs[src]
	if !ok {
		return nil, fmt.Errorf("no ffmpeg download configuration for %s", src)
	}

	cacheDir, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	downloadURL := opts.DownloadURL
	if downloadURL == "" {
		downloadURL = config.ffmpegURL
	}

	destPath := filepath.Join(cacheDir, destBinaries[0])

	if config.isArchive {
		// Download and extract archive.
		err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.ffmpeg, config.ffprobe})
		if err != nil {
			return nil, fmt.Errorf("failed to download and extract ffmpeg archive: %w", err)
		}
	} else {
		// Download single binary.
		destPath, err = downloadFile(ctx, downloadURL, cacheDir, destPath, 0o700)
		if err != nil {
			return nil, fmt.Errorf("failed to download ffmpeg: %w", err)
		}
	}

	return &ResolvedInstall{
		Executable: destPath,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func downloadAndInstallFFprobe(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	src, destBinaries, err := ffprobeGetDownloadBinary()
	if err != nil {
		return nil, err
	}

	config, ok := ffmpegBinConfigs[src]
	if !ok {
		return nil, fmt.Errorf("no ffprobe download configuration for %s", src)
	}

	cacheDir, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(cacheDir, destBinaries[0])

	if config.isArchive {
		// Download and extract archive (contains both ffmpeg and ffprobe).
		downloadURL := opts.DownloadURL
		if downloadURL == "" {
			downloadURL = config.ffmpegURL // Use ffmpeg URL for archive.
		}

		err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.ffmpeg, config.ffprobe})
		if err != nil {
			return nil, fmt.Errorf("failed to download and extract ffprobe archive: %w", err)
		}
	} else {
		// Download single binary.
		downloadURL := opts.DownloadURL
		if downloadURL == "" {
			downloadURL = config.ffprobeURL
		}

		destPath, err = downloadFile(ctx, downloadURL, cacheDir, "", 0o700)
		if err != nil {
			return nil, fmt.Errorf("failed to download ffprobe: %w", err)
		}
	}

	return &ResolvedInstall{
		Executable: destPath,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func ffmpegGetDownloadBinary() (src string, dest []string, err error) {
	src = runtime.GOOS + "_" + runtime.GOARCH
	if binary, ok := ffmpegBinConfigs[src]; ok {
		return src, []string{binary.ffmpeg}, nil
	}

	var supported []string
	for k := range ffmpegBinConfigs {
		supported = append(supported, k)
	}

	if runtime.GOOS == "windows" {
		dest = []string{"ffmpeg.exe"}
	} else {
		dest = []string{"ffmpeg"}
	}

	return src, dest, fmt.Errorf(
		"unsupported os/arch combo: %s/%s (supported: %s)",
		runtime.GOOS,
		runtime.GOARCH,
		strings.Join(supported, ", "),
	)
}

func ffprobeGetDownloadBinary() (src string, dest []string, err error) {
	src = runtime.GOOS + "_" + runtime.GOARCH
	if binary, ok := ffmpegBinConfigs[src]; ok {
		return src, []string{binary.ffprobe}, nil
	}

	var supported []string
	for k := range ffmpegBinConfigs {
		supported = append(supported, k)
	}

	if runtime.GOOS == "windows" {
		dest = []string{"ffprobe.exe"}
	} else {
		dest = []string{"ffprobe"}
	}

	return src, dest, fmt.Errorf(
		"unsupported os/arch combo: %s/%s (supported: %s)",
		runtime.GOOS,
		runtime.GOARCH,
		strings.Join(supported, ", "),
	)
}

var ffmpegVersionRegex = regexp.MustCompile(`^(?:ffmpeg|ffprobe) version ([^ ]+) .*`)

func ffGetVersion(ctx context.Context, r *ResolvedInstall) error {
	var stdout bytes.Buffer

	cmd := exec.Command(r.Executable, "-version") //nolint:gosec
	cmd.Stdout = &stdout
	applySyscall(cmd, false)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run %s to verify version: %w", r.Executable, err)
	}

	version := ffmpegVersionRegex.FindStringSubmatch(stdout.String())
	if len(version) < 2 {
		return fmt.Errorf("unable to parse %s version from output", r.Executable)
	}

	r.Version = version[1]
	debug(ctx, "resolved version", "binary", r.Executable, "version", r.Version)
	return nil
}
