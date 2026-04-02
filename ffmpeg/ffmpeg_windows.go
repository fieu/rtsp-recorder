//go:build windows
// +build windows

package ffmpeg

import (
	"os"
	"os/exec"
	"syscall"
)

// setupProcessGroup configures the process for Windows
// (Windows doesn't have process groups like Unix, but we can set flags)
func setupProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// On Windows, we create a new process group for cleanup
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
}

// killProcessGroup forcefully kills the process on Windows
// (Windows doesn't support negative PIDs like Unix)
func killProcessGroup(pid int) error {
	// Find the process and terminate it
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
