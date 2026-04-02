//go:build !windows
// +build !windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

// setupProcessGroup configures the process to run in a new process group
// (Unix-specific implementation)
func setupProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// killProcessGroup forcefully kills the entire process group
// (Unix-specific implementation using negative PID)
func killProcessGroup(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}
