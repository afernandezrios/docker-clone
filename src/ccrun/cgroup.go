package ccrun

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Creates Container Cgroup (containercg) manually
func NewCgroup() (cgroupPath string) {

	// cgroup location in Ubuntu
	containercg := "/sys/fs/cgroup/containercg"

	if err := os.MkdirAll(containercg, 0755); err != nil {
		fmt.Printf("Error creating cgroup: %v\n", err)
		return
	}

	// Limit max cpu usage to 50% and memory usage to 1000MB
	if err := os.WriteFile(filepath.Join(containercg, "cpu.max"), []byte("50000 100000"), 0700); err != nil {
		fmt.Printf("Error setting cpu limits: %v\n", err)
		return
	}

	if err := os.WriteFile(filepath.Join(containercg, "memory.max"), []byte("1000M"), 0700); err != nil {
		fmt.Printf("Error setting cpu limits: %v\n", err)
		return
	}

	pid := strconv.Itoa(os.Getpid())
	// Add process pid to the cgroup. The cgroup will apply to any child process
	if err := os.WriteFile(filepath.Join(containercg, "cgroup.procs"), []byte(pid), 0700); err != nil {
		fmt.Printf("Error creating cgroup: %v\n", err)
		return
	}

	return containercg
}
