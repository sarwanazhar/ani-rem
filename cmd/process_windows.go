//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// verifyProcess checks if a process with the given PID is running (Windows).
func verifyProcess(pid string) error {
	p, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	// Use tasklist to see if the PID exists
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", p))
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(output), pid) {
		return fmt.Errorf("process not found")
	}
	return nil
}

// killProcess terminates a process with the given PID (Windows).
func killProcess(pid string) error {
	// taskkill /F is the force-kill equivalent
	cmd := exec.Command("taskkill", "/F", "/PID", pid)
	return cmd.Run()
}

// setDetachAttr is a no‑op on Windows; the process already runs
// independently if launched with cmd.Start().
func setDetachAttr(cmd *exec.Cmd) {
	// No special attributes needed for Windows.
}
