//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func verifyProcess(pid string) error {
	p, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
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

func killProcess(pid string) error {
	cmd := exec.Command("taskkill", "/F", "/PID", pid)
	return cmd.Run()
}

func setDetachAttr(cmd *exec.Cmd) {}
