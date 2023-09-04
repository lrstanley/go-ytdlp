// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.
//
// Code generated by cmd/codegen. DO NOT EDIT.
//
// Internet Shortcut Option Group

package ytdlp

// Write an internet shortcut file, depending on the current platform (.url,
// .webloc or .desktop). The URL may be cached by the OS
//
//   - See [Command.UnsetWriteLink], for unsetting the flag.
//   - WriteLink maps to cli flags: --write-link.
func (c *Command) WriteLink() *Command {
	c.addFlag(&Flag{
		ID:   "writelink",
		Flag: "--write-link",
		Args: nil,
	})
	return c
}

// UnsetWriteLink unsets any flags that were previously set by
// [WriteLink].
func (c *Command) UnsetWriteLink() *Command {
	c.removeFlagByID("writelink")
	return c
}

// Write a .url Windows internet shortcut. The OS caches the URL based on the file
// path
//
//   - See [Command.UnsetWriteUrlLink], for unsetting the flag.
//   - WriteUrlLink maps to cli flags: --write-url-link.
func (c *Command) WriteUrlLink() *Command {
	c.addFlag(&Flag{
		ID:   "writeurllink",
		Flag: "--write-url-link",
		Args: nil,
	})
	return c
}

// UnsetWriteUrlLink unsets any flags that were previously set by
// [WriteUrlLink].
func (c *Command) UnsetWriteUrlLink() *Command {
	c.removeFlagByID("writeurllink")
	return c
}

// Write a .webloc macOS internet shortcut
//
//   - See [Command.UnsetWriteWeblocLink], for unsetting the flag.
//   - WriteWeblocLink maps to cli flags: --write-webloc-link.
func (c *Command) WriteWeblocLink() *Command {
	c.addFlag(&Flag{
		ID:   "writewebloclink",
		Flag: "--write-webloc-link",
		Args: nil,
	})
	return c
}

// UnsetWriteWeblocLink unsets any flags that were previously set by
// [WriteWeblocLink].
func (c *Command) UnsetWriteWeblocLink() *Command {
	c.removeFlagByID("writewebloclink")
	return c
}

// Write a .desktop Linux internet shortcut
//
//   - See [Command.UnsetWriteDesktopLink], for unsetting the flag.
//   - WriteDesktopLink maps to cli flags: --write-desktop-link.
func (c *Command) WriteDesktopLink() *Command {
	c.addFlag(&Flag{
		ID:   "writedesktoplink",
		Flag: "--write-desktop-link",
		Args: nil,
	})
	return c
}

// UnsetWriteDesktopLink unsets any flags that were previously set by
// [WriteDesktopLink].
func (c *Command) UnsetWriteDesktopLink() *Command {
	c.removeFlagByID("writedesktoplink")
	return c
}
