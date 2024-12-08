package ccrun

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// Creates Container Cgroup (containercg) manually
func NewCgroup() (cgroupPath string) {

	// cgroup location in Ubuntu
	containercg := "/sys/fs/cgroup/containercg"

	if err := os.MkdirAll(containercg, 0777); err != nil {
		log.Panicf("Error creating cgroup: %v\n", err)
	}

	// Limit max cpu usage to 50%
	if err := os.WriteFile(filepath.Join(containercg, "cpu.max"), []byte("50000 100000"), 0777); err != nil {
		log.Panicf("Error setting cpu limits: %v\n", err)
	}

	// Limit max memory usage to 1000MB
	if err := os.WriteFile(filepath.Join(containercg, "memory.max"), []byte("1000M"), 0777); err != nil {
		log.Panicf("Error setting cpu limits: %v\n", err)
	}

	pid := strconv.Itoa(os.Getpid())
	// Add process pid to the cgroup. The cgroup will apply to any child process
	if err := os.WriteFile(filepath.Join(containercg, "cgroup.procs"), []byte(pid), 0777); err != nil {
		log.Panicf("Error creating cgroup: %v\n", err)
	}

	return containercg
}
