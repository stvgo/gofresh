//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// configureCmdSysProcAttr configura los atributos de proceso específicos para sistemas Unix-like.
func configureCmdSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
