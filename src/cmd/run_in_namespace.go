package cmd

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"fmt"
	"net/http"

	"github.com/afernandezrios/docker-clone/internal/docker"
	"github.com/afernandezrios/docker-clone/internal/os/cgroup"
	"github.com/spf13/cobra"
)

var runInNamespaceCmd = &cobra.Command{
	Use:   "ccrun",
	Short: "ccrun command in container",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return runInNewUTSNamespace(args)
	},
}

func runInNewUTSNamespace(args []string) (e error) {

	if len(args) < 1 {
		return fmt.Errorf("image name is required")
	}

	// Download docker image files (manifest + layers + config)
	downloadDir := "../container-dir/"
	imageName:= args[0]

	dockerClient := docker.New(&http.Client{})

	dockerClient.DownloadImage(downloadDir, imageName)
	
	// Remove all container files when finished
	defer func() {
		if err := os.RemoveAll(downloadDir); err != nil {
			log.Printf("failed to cleanup container dir: %v", err)
		}
	}()

	// Create a cgroup to limit container resources (memory, max cpu, ...)
	cgroupPath := cgroup.New()

	// Rerun same command (/proc/self/exe) in a new namespaces.
	cmd := exec.Command(
		"/proc/self/exe",
		append([]string{"run"}, args[1:]...)...,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &syscall.SysProcAttr{

		Cloneflags:
			syscall.CLONE_NEWUTS | // Isolates hostname and domain name. Allows the container to have its own hostname.
			syscall.CLONE_NEWPID | // Creates a new PID namespace. The child becomes PID 1 inside the container.
			syscall.CLONE_NEWNS |  // Creates a new mount namespace. Allows independent mount/unmount operations.
			syscall.CLONE_NEWCGROUP | // Isolates cgroup membership. Required for proper container cgroup hierarchy.
			syscall.CLONE_NEWUSER, // Creates a new user namespace. Allows “root inside container” without host root privileges.

		Unshareflags: syscall.CLONE_NEWNS, // Forces the current process to unshare the mount namespace before exec. Prevents mount propagation back to the host.

		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,		   // root in container User namespace
				HostID:	  os.Getuid(), // current user UID
				Size:		1,
			},
		},

		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,		   // root in container User namespace
				HostID:	  os.Getgid(), // current user GID
				Size:		1,
			},
		},
	}

	// Start the container process
	if err := cmd.Start(); err != nil {
		log.Printf("failed to start the container process: %v", err)
	}

	// Attach the process to the cgroup
	if err := cgroup.AddProcess(cgroupPath, cmd.Process.Pid); err != nil {
		log.Printf("failed to attach the process to the cgroup: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("failed to wait the container process: %v", err)
	}

	// Cleanup after container exits
	_ = os.RemoveAll(cgroupPath)

	return nil
}
