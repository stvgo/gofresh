//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// configureCmdSysProcAttr configura los atributos de proceso específicos para Windows.
func configureCmdSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NEW_PROCESS_GROUP
	}
}
