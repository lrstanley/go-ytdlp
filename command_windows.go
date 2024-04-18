// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:build windows

package ytdlp

import (
	"os/exec"
	"syscall"
)

// applySyscall applies any OS-specific syscall attributes to the command.
func (c *Command) applySyscall(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW.
		HideWindow:    true,
	}
}
