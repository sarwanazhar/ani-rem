//go:build !windows
// +build !windows

package cmd

import (
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// verifyProcess checks if a process with the given PID is running (Unix).
func verifyProcess(pid string) error {
	p, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	// kill -0 just tests if the process exists
	return syscall.Kill(p, 0)
}

// killProcess terminates a process with the given PID (Unix).
func killProcess(pid string) error {
	p, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	// Try SIGTERM first
	if err := syscall.Kill(p, syscall.SIGTERM); err != nil {
		return err
	}
	// Give it a moment to terminate gracefully
	time.Sleep(500 * time.Millisecond)
	// SIGKILL as a last resort (ignore error – process may already be dead)
	syscall.Kill(p, syscall.SIGKILL)
	return nil
}

// setDetachAttr sets process attributes so the child process survives
// after the parent exits (Unix: Setpgid).
func setDetachAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
