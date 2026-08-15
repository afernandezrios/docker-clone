package vfs

import "syscall"

func Mount() (cleanup func(), err error) {
	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		return nil, err
	}

	cleanup = func() {
		syscall.Unmount("/proc", 0)
	}
	return cleanup, nil
}
