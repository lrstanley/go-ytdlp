// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
)

const (
	xdgCacheDir     = "go-ytdlp"       // Cache directory that will be appended to the XDG cache directory.
	downloadTimeout = 30 * time.Second // HTTP timeout for downloading the yt-dlp binary.
)

var (
	//go:embed ytdlp-public.key
	ytdlpPublicKey []byte // From: https://github.com/yt-dlp/yt-dlp/blob/master/public.key

	resolveCache = atomic.Pointer[ResolvedInstall]{} // Should only be used by [Install].
	installLock  sync.Mutex

	binConfigs = map[string]struct {
		src  string
		dest []string
	}{
		"darwin_amd64":  {"yt-dlp_macos", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"darwin_arm64":  {"yt-dlp_macos", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_amd64":   {"yt-dlp_linux", []string{"yt-dlp-" + Version, "yt-dlp"}},
		"linux_arm64":   {"yt-dlp_linux_aarch64", []string{"yt-dlp-" + Version, "yt-dlp"}},
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

	// DisableChecksum disables checksum verification when downloading.
	DisableChecksum bool

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

func downloadFile(ctx context.Context, url, dest string, perms os.FileMode) error {
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perms)
	if err != nil {
		return fmt.Errorf("unable to create go-ytdlp dependent cache file %q: %w", dest, err)
	}
	defer f.Close()

	// Download the binary.
	client := &http.Client{Timeout: downloadTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to download go-ytdlp dependent file %q: request creation: %w", dest, err)
	}

	req.Header.Set("User-Agent", "github.com/lrstanley/go-ytdlp; version/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to download go-ytdlp dependent file %q: %w", dest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to download go-ytdlp dependent file %q: bad status: %s", dest, resp.Status)
	}

	_, err = f.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to download go-ytdlp dependent file %q: streaming data: %w", dest, err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("unable to download go-ytdlp dependent file %q: closing file: %w", dest, err)
	}

	return nil
}

func githubReleaseAsset(name string) string {
	return fmt.Sprintf("https://github.com/yt-dlp/yt-dlp/releases/download/%s/%s", Version, name)
}

// verifyFileChecksum will verify the checksum of the target file, using the
// checksum file and signature file. If the checksum does not match, an error
// is returned. If the checksum file wasn't signed with the bundled public key,
// an error is also returned.
//
//   - checksum file is expected to be SHA256.
//   - checkAgainst is the name that we should compare to, that will be in the checksum
//     file. If empty, it will use the base name of targetPath.
func verifyFileChecksum(checksumPath, signaturePath, targetPath, checkAgainst string) error {
	if checkAgainst == "" {
		checkAgainst = filepath.Base(targetPath)
	}

	// First validate that the checksum has been properly signed using the known key.
	keyBuf := bytes.NewBuffer(ytdlpPublicKey)

	signatureFile, err := os.Open(signaturePath)
	if err != nil {
		return err
	}
	defer signatureFile.Close()

	checksumFile, err := os.Open(checksumPath)
	if err != nil {
		return err
	}
	defer checksumFile.Close()

	targetFile, err := os.Open(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	keyring, err := openpgp.ReadArmoredKeyRing(keyBuf)
	if err != nil {
		return fmt.Errorf("unable to read armored key ring: %w", err)
	}

	_, err = openpgp.CheckDetachedSignature(keyring, checksumFile, signatureFile, nil)
	if err != nil {
		return fmt.Errorf("unable to check detached signature: %w", err)
	}

	// Now make sure the checksum from checksumFile matches the target file.
	hash := sha256.New()
	if _, err = io.Copy(hash, targetFile); err != nil {
		return err
	}

	sum := hex.EncodeToString(hash.Sum(nil))

	_, err = checksumFile.Seek(0, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(checksumFile)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) != 2 { //nolint:gomnd
			continue
		}

		if fields[1] == checkAgainst {
			if fields[0] != sum {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", fields[0], sum)
			}

			return nil
		}
	}

	return fmt.Errorf("unable to find checksum for %s", filepath.Base(targetPath))
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
		return nil, errors.New("yt-dlp executable not found, and downloading is disabled")
	}

	src, dest, err := getDownloadBinary()
	if err != nil {
		return nil, err
	}

	downloadURL := opts.DownloadURL

	if downloadURL == "" {
		downloadURL = githubReleaseAsset(src)
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

	err = downloadFile(ctx, downloadURL, filepath.Join(dir, dest[0]+".tmp"), 0o750) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	if !opts.DisableChecksum {
		err = downloadFile(ctx, githubReleaseAsset("SHA2-256SUMS"), filepath.Join(dir, "SHA2-256SUMS-"+Version), 0o640) //nolint:gomnd
		if err != nil {
			return nil, err
		}

		err = downloadFile(ctx, githubReleaseAsset("SHA2-256SUMS.sig"), filepath.Join(dir, "SHA2-256SUMS-"+Version+".sig"), 0o640) //nolint:gomnd
		if err != nil {
			return nil, err
		}

		err = verifyFileChecksum(
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
	resolved, err = resolveExecutable(false, true)
	if err != nil {
		return nil, err
	}

	resolveCache.Store(resolved)
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
			if err != nil {
				continue
			}

			if !stat.IsDir() && isExecutable(bin, stat) {
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
