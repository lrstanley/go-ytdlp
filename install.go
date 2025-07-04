// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ulikunitz/xz"
)

const (
	xdgCacheDir     = "go-ytdlp"       // Cache directory that will be appended to the XDG cache directory.
	downloadTimeout = 30 * time.Second // HTTP timeout for downloading the yt-dlp binary.
)

// GetCacheDir returns the cache directory for go-ytdlp. Note that it may not be created yet.
func GetCacheDir() (string, error) {
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine cache directory: %w", err)
	}
	return filepath.Join(baseCacheDir, xdgCacheDir), nil
}

// RemoveInstallCache removes the cache directory for go-ytdlp, and clears in-memory
// install resolve caches for all binaries.
func RemoveInstallCache() error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	debug(context.Background(), "removed cache directory", "path", cacheDir)
	err = os.RemoveAll(cacheDir)
	if err != nil {
		return fmt.Errorf("unable to remove cache directory: %w", err)
	}

	ytdlpResolveCache.Store(nil)
	ffmpegResolveCache.Store(nil)
	ffprobeResolveCache.Store(nil)
	return nil
}

// createCacheDir creates the go-ytdlp cache directory and returns its path.
func createCacheDir(ctx context.Context) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		debug(ctx, "cache directory does not exist, creating", "path", cacheDir)
	}

	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", fmt.Errorf("unable to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// MustInstallAll is similar to [InstallAll], but panics if there is an error.
func MustInstallAll(ctx context.Context) []*ResolvedInstall {
	installs, err := InstallAll(ctx)
	if err != nil {
		panic(err)
	}
	return installs
}

// InstallAll installs all dependencies for go-ytdlp concurrently, using default options.
// Note that this will not work on all platforms, as some dependencies are only supported on
// certain platforms.
func InstallAll(ctx context.Context) ([]*ResolvedInstall, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	_, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	var installs []*ResolvedInstall
	var errs []error

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := Install(ctx, nil)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}
		mu.Lock()
		installs = append(installs, r)
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := InstallFFmpeg(ctx, nil)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}
		mu.Lock()
		installs = append(installs, r)
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := InstallFFprobe(ctx, nil)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}
		mu.Lock()
		installs = append(installs, r)
		mu.Unlock()
	}()

	wg.Wait()

	if len(errs) > 0 {
		return installs, errors.Join(errs...)
	}

	return installs, nil
}

func downloadFile(ctx context.Context, url, targetDir, targetName string, perms os.FileMode) (dest string, err error) {
	debug(
		ctx, "downloading file",
		"url", url,
		"dir", targetDir,
		"file", targetName,
	)

	// Download the binary.
	client := &http.Client{Timeout: downloadTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("unable to download go-ytdlp dependent file %q: request creation: %w", dest, err)
	}

	req.Header.Set("User-Agent", "github.com/lrstanley/go-ytdlp; version/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to download go-ytdlp dependent file %q: %w", dest, err)
	}
	defer resp.Body.Close()

	debug(ctx, "received response", "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to download go-ytdlp dependent file %q: bad status: %s", dest, resp.Status)
	}

	if targetName != "" {
		dest = targetName
	} else {
		disposition, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err != nil {
			return "", fmt.Errorf("unable to parse content disposition: %w", err)
		}
		if disposition != "attachment" {
			return "", fmt.Errorf("unexpected content disposition: %s", disposition)
		}

		if params["filename"] == "" {
			return "", fmt.Errorf("no filename in content disposition")
		}

		dest = filepath.Join(targetDir, params["filename"])
		debug(ctx, "using filename from content disposition", "dest", dest)
	}

	debug(ctx, "creating file", "dest", dest)

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perms)
	if err != nil {
		return "", fmt.Errorf("unable to create go-ytdlp dependent cache file %q: %w", dest, err)
	}
	defer f.Close()

	_, err = f.ReadFrom(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to download go-ytdlp dependent file %q: streaming data: %w", dest, err)
	}

	err = f.Close()
	if err != nil {
		return "", fmt.Errorf("unable to download go-ytdlp dependent file %q: closing file: %w", dest, err)
	}

	return dest, nil
}

// downloadAndExtractFilesFromArchive downloads an archive from the given URL, extracts the specified files
// into cacheDir, and removes the archive after extraction.
func downloadAndExtractFilesFromArchive(ctx context.Context, downloadURL, cacheDir string, filenames []string) error {
	dest, err := downloadFile(ctx, downloadURL, cacheDir, "", 0o644)
	if err != nil {
		return err
	}
	defer os.Remove(dest)

	return extractFilesFromArchive(ctx, dest, cacheDir, filenames)
}

// extractFilesFromArchive extracts the specified files from the given archive (zip, tar.xz) into cacheDir.
// The archive type is detected from the file extension.
func extractFilesFromArchive(ctx context.Context, archivePath, cacheDir string, filenames []string) error {
	var archiveType string
	if strings.HasSuffix(archivePath, ".zip") {
		archiveType = "zip"
	} else if strings.HasSuffix(archivePath, ".tar.xz") {
		archiveType = "tar.xz"
	} else {
		return fmt.Errorf("unsupported archive format for file: %s", archivePath)
	}

	switch archiveType {
	case "zip":
		debug(
			ctx, "extracting zip archive",
			"archive", archivePath,
			"cache", cacheDir,
			"filenames", filenames,
		)

		reader, err := zip.OpenReader(archivePath)
		if err != nil {
			return err
		}
		defer reader.Close()

		for _, file := range reader.File {
			for _, name := range filenames {
				if strings.HasSuffix(file.Name, "/"+name) || strings.HasSuffix(file.Name, "\\"+name) || file.Name == name {
					rc, err := file.Open()
					if err != nil {
						return err
					}
					defer rc.Close()

					debug(
						ctx, "extracting file",
						"archive", archivePath,
						"cache", cacheDir,
						"filenames", filenames,
						"name", name,
					)

					outFile, err := os.OpenFile(filepath.Join(cacheDir, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
					if err != nil {
						return err
					}
					defer outFile.Close()

					_, err = io.Copy(outFile, rc)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	case "tar.xz":
		debug(
			ctx, "extracting tar.xz archive",
			"archive", archivePath,
			"cache", cacheDir,
			"filenames", filenames,
		)

		file, err := os.Open(archivePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// ref: https://github.com/hashicorp/go-getter/pull/520
		xzReader, err := xz.NewReader(bufio.NewReader(file))
		if err != nil {
			return err
		}

		tarReader := tar.NewReader(xzReader)

		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			for _, name := range filenames {
				if strings.HasSuffix(header.Name, "/"+name) || header.Name == name {
					debug(ctx, "extracting file", "archivePath", archivePath, "cacheDir", cacheDir, "filenames", filenames, "name", name)

					outFile, err := os.OpenFile(filepath.Join(cacheDir, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
					if err != nil {
						return err
					}
					defer outFile.Close()

					_, err = io.Copy(outFile, tarReader)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported archive type: %s", archiveType)
	}
}

// verifyFileChecksum will verify the checksum of the target file, using the
// checksum file and signature file. If the checksum does not match, an error
// is returned. If the checksum file wasn't signed with the bundled public key,
// an error is also returned.
//
//   - checksum file is expected to be SHA256.
//   - checkAgainst is the name that we should compare to, that will be in the checksum
//     file. If empty, it will use the base name of targetPath.
func verifyFileChecksum(ctx context.Context, checksumPath, signaturePath, targetPath, checkAgainst string) error {
	if checkAgainst == "" {
		checkAgainst = filepath.Base(targetPath)
	}

	debug(
		ctx, "verifying file checksum",
		"checksum", checksumPath,
		"signature", signaturePath,
		"target", targetPath,
		"against", checkAgainst,
	)

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

// ResolvedInstall is the found executable.
type ResolvedInstall struct {
	Executable string // Path to the executable.
	Version    string // Version that was resolved. If [InstallOptions.AllowVersionMismatch] is specified, this will be empty.
	FromCache  bool   // Whether the executable was resolved from the cache.
	Downloaded bool   // Whether the executable was downloaded during this invocation.
}

// resolveExecutable will attempt to resolve the yt-dlp executable, either from
// the go-ytdlp cache (first), or from the PATH (second). If it's not found, an
// error is returned.
func resolveExecutable(ctx context.Context, calleeIsDownloader, disableSystem bool, binaries []string) (r *ResolvedInstall, err error) {
	var stat os.FileInfo
	var bin, baseCacheDir string

	baseCacheDir, err = os.UserCacheDir()
	if err == nil {
		// Check out cache dirs first.
		for _, d := range binaries {
			bin = filepath.Join(baseCacheDir, xdgCacheDir, d)

			stat, err = os.Stat(bin)
			if err != nil {
				continue
			}

			if !stat.IsDir() && isExecutable(bin, stat) {
				debug(ctx, "found executable in cache", "path", bin)

				r = &ResolvedInstall{
					Executable: bin,
					FromCache:  true,
					Downloaded: calleeIsDownloader,
				}
				if calleeIsDownloader {
					r.Version = Version
				}
				return r, nil
			}
		}
	}

	if !disableSystem {
		// Check PATH for the binary.
		for _, d := range binaries {
			bin, err = exec.LookPath(d)
			if err == nil {
				debug(ctx, "found executable in PATH", "path", bin)

				return &ResolvedInstall{
					Executable: bin,
					FromCache:  false,
					Downloaded: false,
				}, nil
			}
		}
	}

	// Will pick the last error, which is likely without the version suffix, what we want.
	return nil, fmt.Errorf("unable to resolve executable from provided paths (%s): %w", strings.Join(binaries, ", "), err)
}
