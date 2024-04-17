// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:build !windows

package ytdlp

import (
	"os/exec"
)

// applySyscall applies any OS-specific syscall attributes to the command.
func (c *Command) applySyscall(_ *exec.Cmd) {
	// No-op by default.
}
