//go:build !windows
// +build !windows

package cmd

import (
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func verifyProcess(pid string) error {
	p, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	return syscall.Kill(p, 0)
}

func killProcess(pid string) error {
	p, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	if err := syscall.Kill(p, syscall.SIGTERM); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	syscall.Kill(p, syscall.SIGKILL)
	return nil
}

func setDetachAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
