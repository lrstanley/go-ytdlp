// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.
//
// Code generated by cmd/codegen. DO NOT EDIT.
//
// Filesystem Option Group

package ytdlp

import (
	"strconv"
)

// File containing URLs to download ("-" for stdin), one URL per line. Lines
// starting with "#", ";" or "]" are considered as comments and ignored
//
//   - See [UnsetBatchFile], for unsetting the flag.
//   - BatchFile maps to cli flags: -a/--batch-file=FILE.
func (c *Command) BatchFile(file string) *Command {
	c.addFlag(&Flag{
		ID:   "batchfile",
		Flag: "--batch-file",
		Args: []string{file},
	})
	return c
}

// UnsetBatchFile unsets any flags that were previously set by
// [BatchFile].
func (c *Command) UnsetBatchFile() *Command {
	c.removeFlagByID("batchfile")
	return c
}

// Do not read URLs from batch file (default)
//
//   - See [UnsetBatchFile], for unsetting the flag.
//   - NoBatchFile maps to cli flags: --no-batch-file.
func (c *Command) NoBatchFile() *Command {
	c.addFlag(&Flag{
		ID:   "batchfile",
		Flag: "--no-batch-file",
		Args: nil,
	})
	return c
}

// Id sets the "id" flag (no description specified).
//
//   - See [UnsetId], for unsetting the flag.
//   - Id maps to cli flags: --id (hidden).
func (c *Command) Id() *Command {
	c.addFlag(&Flag{
		ID:   "useid",
		Flag: "--id",
		Args: nil,
	})
	return c
}

// UnsetId unsets any flags that were previously set by
// [Id].
func (c *Command) UnsetId() *Command {
	c.removeFlagByID("useid")
	return c
}

// The paths where the files should be downloaded. Specify the type of file and the
// path separated by a colon ":". All the same TYPES as --output are supported.
// Additionally, you can also provide "home" (default) and "temp" paths. All
// intermediary files are first downloaded to the temp path and then the final
// files are moved over to the home path after download is finished. This option is
// ignored if --output is an absolute path
//
//   - See [UnsetPaths], for unsetting the flag.
//   - Paths maps to cli flags: -P/--paths=[TYPES:]PATH.
func (c *Command) Paths(path string) *Command {
	c.addFlag(&Flag{
		ID:   "paths",
		Flag: "--paths",
		Args: []string{path},
	})
	return c
}

// UnsetPaths unsets any flags that were previously set by
// [Paths].
func (c *Command) UnsetPaths() *Command {
	c.removeFlagByID("paths")
	return c
}

// Output filename template; see "OUTPUT TEMPLATE" for details
//
//   - See [UnsetOutput], for unsetting the flag.
//   - Output maps to cli flags: -o/--output=[TYPES:]TEMPLATE.
func (c *Command) Output(template string) *Command {
	c.addFlag(&Flag{
		ID:   "outtmpl",
		Flag: "--output",
		Args: []string{template},
	})
	return c
}

// UnsetOutput unsets any flags that were previously set by
// [Output].
func (c *Command) UnsetOutput() *Command {
	c.removeFlagByID("outtmpl")
	return c
}

// Placeholder for unavailable fields in "OUTPUT TEMPLATE" (default: "NA")
//
//   - See [UnsetOutputNaPlaceholder], for unsetting the flag.
//   - OutputNaPlaceholder maps to cli flags: --output-na-placeholder=TEXT.
func (c *Command) OutputNaPlaceholder(text string) *Command {
	c.addFlag(&Flag{
		ID:   "outtmpl_na_placeholder",
		Flag: "--output-na-placeholder",
		Args: []string{text},
	})
	return c
}

// UnsetOutputNaPlaceholder unsets any flags that were previously set by
// [OutputNaPlaceholder].
func (c *Command) UnsetOutputNaPlaceholder() *Command {
	c.removeFlagByID("outtmpl_na_placeholder")
	return c
}

// - See [UnsetAutonumberSize], for unsetting the flag.
// - AutonumberSize maps to cli flags: --autonumber-size=NUMBER (hidden).
func (c *Command) AutonumberSize(number int) *Command {
	c.addFlag(&Flag{
		ID:   "autonumber_size",
		Flag: "--autonumber-size",
		Args: []string{
			strconv.Itoa(number),
		},
	})
	return c
}

// UnsetAutonumberSize unsets any flags that were previously set by
// [AutonumberSize].
func (c *Command) UnsetAutonumberSize() *Command {
	c.removeFlagByID("autonumber_size")
	return c
}

// - See [UnsetAutonumberStart], for unsetting the flag.
// - AutonumberStart maps to cli flags: --autonumber-start=NUMBER (hidden).
func (c *Command) AutonumberStart(number int) *Command {
	c.addFlag(&Flag{
		ID:   "autonumber_start",
		Flag: "--autonumber-start",
		Args: []string{
			strconv.Itoa(number),
		},
	})
	return c
}

// UnsetAutonumberStart unsets any flags that were previously set by
// [AutonumberStart].
func (c *Command) UnsetAutonumberStart() *Command {
	c.removeFlagByID("autonumber_start")
	return c
}

// Restrict filenames to only ASCII characters, and avoid "&" and spaces in
// filenames
//
//   - See [UnsetRestrictFilenames], for unsetting the flag.
//   - RestrictFilenames maps to cli flags: --restrict-filenames.
func (c *Command) RestrictFilenames() *Command {
	c.addFlag(&Flag{
		ID:   "restrictfilenames",
		Flag: "--restrict-filenames",
		Args: nil,
	})
	return c
}

// UnsetRestrictFilenames unsets any flags that were previously set by
// [RestrictFilenames].
func (c *Command) UnsetRestrictFilenames() *Command {
	c.removeFlagByID("restrictfilenames")
	return c
}

// Allow Unicode characters, "&" and spaces in filenames (default)
//
//   - See [UnsetRestrictFilenames], for unsetting the flag.
//   - NoRestrictFilenames maps to cli flags: --no-restrict-filenames.
func (c *Command) NoRestrictFilenames() *Command {
	c.addFlag(&Flag{
		ID:   "restrictfilenames",
		Flag: "--no-restrict-filenames",
		Args: nil,
	})
	return c
}

// Force filenames to be Windows-compatible
//
//   - See [UnsetWindowsFilenames], for unsetting the flag.
//   - WindowsFilenames maps to cli flags: --windows-filenames.
func (c *Command) WindowsFilenames() *Command {
	c.addFlag(&Flag{
		ID:   "windowsfilenames",
		Flag: "--windows-filenames",
		Args: nil,
	})
	return c
}

// UnsetWindowsFilenames unsets any flags that were previously set by
// [WindowsFilenames].
func (c *Command) UnsetWindowsFilenames() *Command {
	c.removeFlagByID("windowsfilenames")
	return c
}

// Make filenames Windows-compatible only if using Windows (default)
//
//   - See [UnsetWindowsFilenames], for unsetting the flag.
//   - NoWindowsFilenames maps to cli flags: --no-windows-filenames.
func (c *Command) NoWindowsFilenames() *Command {
	c.addFlag(&Flag{
		ID:   "windowsfilenames",
		Flag: "--no-windows-filenames",
		Args: nil,
	})
	return c
}

// Limit the filename length (excluding extension) to the specified number of
// characters
//
//   - See [UnsetTrimFilenames], for unsetting the flag.
//   - TrimFilenames maps to cli flags: --trim-filenames/--trim-file-names=LENGTH.
func (c *Command) TrimFilenames(length int) *Command {
	c.addFlag(&Flag{
		ID:   "trim_file_name",
		Flag: "--trim-filenames",
		Args: []string{
			strconv.Itoa(length),
		},
	})
	return c
}

// UnsetTrimFilenames unsets any flags that were previously set by
// [TrimFilenames].
func (c *Command) UnsetTrimFilenames() *Command {
	c.removeFlagByID("trim_file_name")
	return c
}

// Do not overwrite any files
//
//   - See [UnsetOverwrites], for unsetting the flag.
//   - NoOverwrites maps to cli flags: -w/--no-overwrites.
func (c *Command) NoOverwrites() *Command {
	c.addFlag(&Flag{
		ID:   "overwrites",
		Flag: "--no-overwrites",
		Args: nil,
	})
	return c
}

// Overwrite all video and metadata files. This option includes --no-continue
//
//   - See [UnsetForceOverwrites], for unsetting the flag.
//   - ForceOverwrites maps to cli flags: --force-overwrites/--yes-overwrites.
func (c *Command) ForceOverwrites() *Command {
	c.addFlag(&Flag{
		ID:   "overwrites",
		Flag: "--force-overwrites",
		Args: nil,
	})
	return c
}

// UnsetForceOverwrites unsets any flags that were previously set by
// [ForceOverwrites].
func (c *Command) UnsetForceOverwrites() *Command {
	c.removeFlagByID("overwrites")
	return c
}

// Do not overwrite the video, but overwrite related files (default)
//
//   - See [UnsetForceOverwrites], for unsetting the flag.
//   - NoForceOverwrites maps to cli flags: --no-force-overwrites.
func (c *Command) NoForceOverwrites() *Command {
	c.addFlag(&Flag{
		ID:   "overwrites",
		Flag: "--no-force-overwrites",
		Args: nil,
	})
	return c
}

// Resume partially downloaded files/fragments (default)
//
//   - See [UnsetContinue], for unsetting the flag.
//   - Continue maps to cli flags: -c/--continue.
func (c *Command) Continue() *Command {
	c.addFlag(&Flag{
		ID:   "continue_dl",
		Flag: "--continue",
		Args: nil,
	})
	return c
}

// UnsetContinue unsets any flags that were previously set by
// [Continue].
func (c *Command) UnsetContinue() *Command {
	c.removeFlagByID("continue_dl")
	return c
}

// Do not resume partially downloaded fragments. If the file is not fragmented,
// restart download of the entire file
//
//   - See [UnsetContinue], for unsetting the flag.
//   - NoContinue maps to cli flags: --no-continue.
func (c *Command) NoContinue() *Command {
	c.addFlag(&Flag{
		ID:   "continue_dl",
		Flag: "--no-continue",
		Args: nil,
	})
	return c
}

// Use .part files instead of writing directly into output file (default)
//
//   - See [UnsetPart], for unsetting the flag.
//   - Part maps to cli flags: --part.
func (c *Command) Part() *Command {
	c.addFlag(&Flag{
		ID:   "nopart",
		Flag: "--part",
		Args: nil,
	})
	return c
}

// UnsetPart unsets any flags that were previously set by
// [Part].
func (c *Command) UnsetPart() *Command {
	c.removeFlagByID("nopart")
	return c
}

// Do not use .part files - write directly into output file
//
//   - See [UnsetPart], for unsetting the flag.
//   - NoPart maps to cli flags: --no-part.
func (c *Command) NoPart() *Command {
	c.addFlag(&Flag{
		ID:   "nopart",
		Flag: "--no-part",
		Args: nil,
	})
	return c
}

// Use the Last-modified header to set the file modification time (default)
//
//   - See [UnsetMtime], for unsetting the flag.
//   - Mtime maps to cli flags: --mtime.
func (c *Command) Mtime() *Command {
	c.addFlag(&Flag{
		ID:   "updatetime",
		Flag: "--mtime",
		Args: nil,
	})
	return c
}

// UnsetMtime unsets any flags that were previously set by
// [Mtime].
func (c *Command) UnsetMtime() *Command {
	c.removeFlagByID("updatetime")
	return c
}

// Do not use the Last-modified header to set the file modification time
//
//   - See [UnsetMtime], for unsetting the flag.
//   - NoMtime maps to cli flags: --no-mtime.
func (c *Command) NoMtime() *Command {
	c.addFlag(&Flag{
		ID:   "updatetime",
		Flag: "--no-mtime",
		Args: nil,
	})
	return c
}

// Write video description to a .description file
//
//   - See [UnsetWriteDescription], for unsetting the flag.
//   - WriteDescription maps to cli flags: --write-description.
func (c *Command) WriteDescription() *Command {
	c.addFlag(&Flag{
		ID:   "writedescription",
		Flag: "--write-description",
		Args: nil,
	})
	return c
}

// UnsetWriteDescription unsets any flags that were previously set by
// [WriteDescription].
func (c *Command) UnsetWriteDescription() *Command {
	c.removeFlagByID("writedescription")
	return c
}

// Do not write video description (default)
//
//   - See [UnsetWriteDescription], for unsetting the flag.
//   - NoWriteDescription maps to cli flags: --no-write-description.
func (c *Command) NoWriteDescription() *Command {
	c.addFlag(&Flag{
		ID:   "writedescription",
		Flag: "--no-write-description",
		Args: nil,
	})
	return c
}

// Write video metadata to a .info.json file (this may contain personal
// information)
//
//   - See [UnsetWriteInfoJson], for unsetting the flag.
//   - WriteInfoJson maps to cli flags: --write-info-json.
func (c *Command) WriteInfoJson() *Command {
	c.addFlag(&Flag{
		ID:   "writeinfojson",
		Flag: "--write-info-json",
		Args: nil,
	})
	return c
}

// UnsetWriteInfoJson unsets any flags that were previously set by
// [WriteInfoJson].
func (c *Command) UnsetWriteInfoJson() *Command {
	c.removeFlagByID("writeinfojson")
	return c
}

// Do not write video metadata (default)
//
//   - See [UnsetWriteInfoJson], for unsetting the flag.
//   - NoWriteInfoJson maps to cli flags: --no-write-info-json.
func (c *Command) NoWriteInfoJson() *Command {
	c.addFlag(&Flag{
		ID:   "writeinfojson",
		Flag: "--no-write-info-json",
		Args: nil,
	})
	return c
}

// WriteAnnotations sets the "write-annotations" flag (no description specified).
//
//   - See [UnsetWriteAnnotations], for unsetting the flag.
//   - WriteAnnotations maps to cli flags: --write-annotations (hidden).
func (c *Command) WriteAnnotations() *Command {
	c.addFlag(&Flag{
		ID:   "writeannotations",
		Flag: "--write-annotations",
		Args: nil,
	})
	return c
}

// UnsetWriteAnnotations unsets any flags that were previously set by
// [WriteAnnotations].
func (c *Command) UnsetWriteAnnotations() *Command {
	c.removeFlagByID("writeannotations")
	return c
}

// NoWriteAnnotations sets the "no-write-annotations" flag (no description specified).
//
//   - See [UnsetWriteAnnotations], for unsetting the flag.
//   - NoWriteAnnotations maps to cli flags: --no-write-annotations (hidden).
func (c *Command) NoWriteAnnotations() *Command {
	c.addFlag(&Flag{
		ID:   "writeannotations",
		Flag: "--no-write-annotations",
		Args: nil,
	})
	return c
}

// Write playlist metadata in addition to the video metadata when using
// --write-info-json, --write-description etc. (default)
//
//   - See [UnsetWritePlaylistMetafiles], for unsetting the flag.
//   - WritePlaylistMetafiles maps to cli flags: --write-playlist-metafiles.
func (c *Command) WritePlaylistMetafiles() *Command {
	c.addFlag(&Flag{
		ID:   "allow_playlist_files",
		Flag: "--write-playlist-metafiles",
		Args: nil,
	})
	return c
}

// UnsetWritePlaylistMetafiles unsets any flags that were previously set by
// [WritePlaylistMetafiles].
func (c *Command) UnsetWritePlaylistMetafiles() *Command {
	c.removeFlagByID("allow_playlist_files")
	return c
}

// Do not write playlist metadata when using --write-info-json, --write-description
// etc.
//
//   - See [UnsetWritePlaylistMetafiles], for unsetting the flag.
//   - NoWritePlaylistMetafiles maps to cli flags: --no-write-playlist-metafiles.
func (c *Command) NoWritePlaylistMetafiles() *Command {
	c.addFlag(&Flag{
		ID:   "allow_playlist_files",
		Flag: "--no-write-playlist-metafiles",
		Args: nil,
	})
	return c
}

// Remove some internal metadata such as filenames from the infojson (default)
//
//   - See [UnsetCleanInfoJson], for unsetting the flag.
//   - CleanInfoJson maps to cli flags: --clean-info-json/--clean-infojson.
func (c *Command) CleanInfoJson() *Command {
	c.addFlag(&Flag{
		ID:   "clean_infojson",
		Flag: "--clean-info-json",
		Args: nil,
	})
	return c
}

// UnsetCleanInfoJson unsets any flags that were previously set by
// [CleanInfoJson].
func (c *Command) UnsetCleanInfoJson() *Command {
	c.removeFlagByID("clean_infojson")
	return c
}

// Write all fields to the infojson
//
//   - See [UnsetCleanInfoJson], for unsetting the flag.
//   - NoCleanInfoJson maps to cli flags: --no-clean-info-json/--no-clean-infojson.
func (c *Command) NoCleanInfoJson() *Command {
	c.addFlag(&Flag{
		ID:   "clean_infojson",
		Flag: "--no-clean-info-json",
		Args: nil,
	})
	return c
}

// Retrieve video comments to be placed in the infojson. The comments are fetched
// even without this option if the extraction is known to be quick (Alias:
// --get-comments)
//
//   - See [UnsetWriteComments], for unsetting the flag.
//   - WriteComments maps to cli flags: --write-comments/--get-comments.
func (c *Command) WriteComments() *Command {
	c.addFlag(&Flag{
		ID:   "getcomments",
		Flag: "--write-comments",
		Args: nil,
	})
	return c
}

// UnsetWriteComments unsets any flags that were previously set by
// [WriteComments].
func (c *Command) UnsetWriteComments() *Command {
	c.removeFlagByID("getcomments")
	return c
}

// Do not retrieve video comments unless the extraction is known to be quick
// (Alias: --no-get-comments)
//
//   - See [UnsetWriteComments], for unsetting the flag.
//   - NoWriteComments maps to cli flags: --no-write-comments/--no-get-comments.
func (c *Command) NoWriteComments() *Command {
	c.addFlag(&Flag{
		ID:   "getcomments",
		Flag: "--no-write-comments",
		Args: nil,
	})
	return c
}

// JSON file containing the video information (created with the "--write-info-json"
// option)
//
//   - See [UnsetLoadInfoJson], for unsetting the flag.
//   - LoadInfoJson maps to cli flags: --load-info-json/--load-info=FILE.
func (c *Command) LoadInfoJson(file string) *Command {
	c.addFlag(&Flag{
		ID:   "load_info_filename",
		Flag: "--load-info-json",
		Args: []string{file},
	})
	return c
}

// UnsetLoadInfoJson unsets any flags that were previously set by
// [LoadInfoJson].
func (c *Command) UnsetLoadInfoJson() *Command {
	c.removeFlagByID("load_info_filename")
	return c
}

// Netscape formatted file to read cookies from and dump cookie jar in
//
//   - See [UnsetCookies], for unsetting the flag.
//   - Cookies maps to cli flags: --cookies=FILE.
func (c *Command) Cookies(file string) *Command {
	c.addFlag(&Flag{
		ID:   "cookiefile",
		Flag: "--cookies",
		Args: []string{file},
	})
	return c
}

// UnsetCookies unsets any flags that were previously set by
// [Cookies].
func (c *Command) UnsetCookies() *Command {
	c.removeFlagByID("cookiefile")
	return c
}

// Do not read/dump cookies from/to file (default)
//
//   - See [UnsetCookies], for unsetting the flag.
//   - NoCookies maps to cli flags: --no-cookies=FILE.
func (c *Command) NoCookies() *Command {
	c.addFlag(&Flag{
		ID:   "cookiefile",
		Flag: "--no-cookies",
		Args: nil,
	})
	return c
}

// The name of the browser to load cookies from. Currently supported browsers are:
// brave, chrome, chromium, edge, firefox, opera, safari, vivaldi. Optionally, the
// KEYRING used for decrypting Chromium cookies on Linux, the name/path of the
// PROFILE to load cookies from, and the CONTAINER name (if Firefox) ("none" for no
// container) can be given with their respective seperators. By default, all
// containers of the most recently accessed profile are used. Currently supported
// keyrings are: basictext, gnomekeyring, kwallet, kwallet5, kwallet6
//
//   - See [UnsetCookiesFromBrowser], for unsetting the flag.
//   - CookiesFromBrowser maps to cli flags: --cookies-from-browser=BROWSER[+KEYRING][:PROFILE][::CONTAINER].
func (c *Command) CookiesFromBrowser(browser string) *Command {
	c.addFlag(&Flag{
		ID:   "cookiesfrombrowser",
		Flag: "--cookies-from-browser",
		Args: []string{browser},
	})
	return c
}

// UnsetCookiesFromBrowser unsets any flags that were previously set by
// [CookiesFromBrowser].
func (c *Command) UnsetCookiesFromBrowser() *Command {
	c.removeFlagByID("cookiesfrombrowser")
	return c
}

// Do not load cookies from browser (default)
//
//   - See [UnsetCookiesFromBrowser], for unsetting the flag.
//   - NoCookiesFromBrowser maps to cli flags: --no-cookies-from-browser.
func (c *Command) NoCookiesFromBrowser() *Command {
	c.addFlag(&Flag{
		ID:   "cookiesfrombrowser",
		Flag: "--no-cookies-from-browser",
		Args: nil,
	})
	return c
}

// Location in the filesystem where yt-dlp can store some downloaded information
// (such as client ids and signatures) permanently. By default
// ${XDG_CACHE_HOME}/yt-dlp
//
//   - See [UnsetCacheDir], for unsetting the flag.
//   - CacheDir maps to cli flags: --cache-dir=DIR.
func (c *Command) CacheDir(dir string) *Command {
	c.addFlag(&Flag{
		ID:   "cachedir",
		Flag: "--cache-dir",
		Args: []string{dir},
	})
	return c
}

// UnsetCacheDir unsets any flags that were previously set by
// [CacheDir].
func (c *Command) UnsetCacheDir() *Command {
	c.removeFlagByID("cachedir")
	return c
}

// Disable filesystem caching
//
//   - See [UnsetCacheDir], for unsetting the flag.
//   - NoCacheDir maps to cli flags: --no-cache-dir.
func (c *Command) NoCacheDir() *Command {
	c.addFlag(&Flag{
		ID:   "cachedir",
		Flag: "--no-cache-dir",
		Args: nil,
	})
	return c
}

// Delete all filesystem cache files
//
//   - See [UnsetRmCacheDir], for unsetting the flag.
//   - RmCacheDir maps to cli flags: --rm-cache-dir.
func (c *Command) RmCacheDir() *Command {
	c.addFlag(&Flag{
		ID:   "rm_cachedir",
		Flag: "--rm-cache-dir",
		Args: nil,
	})
	return c
}

// UnsetRmCacheDir unsets any flags that were previously set by
// [RmCacheDir].
func (c *Command) UnsetRmCacheDir() *Command {
	c.removeFlagByID("rm_cachedir")
	return c
}
