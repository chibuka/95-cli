//go:build !windows

package runner

import "syscall"

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func killProcess(pid int) error {
	return syscall.Kill(-pid, syscall.SIGTERM)
}
