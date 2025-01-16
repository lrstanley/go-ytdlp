// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:build unix

package ytdlp

import (
	"os/exec"
	"syscall"
)

// applySyscall applies any OS-specific syscall attributes to the command.
func (c *Command) applySyscall(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: c.separateProcessGroup,
	}
}
