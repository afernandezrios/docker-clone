package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/afernandezrios/docker-clone/internal/docker"
	"github.com/afernandezrios/docker-clone/internal/os/cgroup"
	"github.com/spf13/cobra"
)

var runInNamespaceCmd = &cobra.Command{
	Use:   "ccrun <image> <command> [args...]",
	Short: "Run a container command in isolated Linux namespaces",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]
		containerCmd := args[1:]

		return launchContainer(image, containerCmd)
	},
}

func launchContainer(image string, command []string) error {

	// Prepare isolated root filesystem
	tempDir, err := os.MkdirTemp("", "container-rootfs-*")
	if err != nil {
		return fmt.Errorf("create temp rootfs dir: %w", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(tempDir); rmErr != nil {
			fmt.Fprintf(os.Stderr, "failed to cleanup temp rootfs %s: %v\n", tempDir, rmErr)
		}
	}()

	client := docker.New(&http.Client{})
	if err := client.DownloadImage(image, tempDir); err != nil {
		return fmt.Errorf("download image %q: %w", image, err)
	}

	// Remove all container files when finished
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed to cleanup container dir %s: %v\n", tempDir, err)
		}
	}()

	// Create a cgroup to limit container resources (memory, max cpu, ...)
	cgroupPath := cgroup.New()

	// Rerun same command (/proc/self/exe) in a new namespaces.
	execArgs := append([]string{"run", "--rootfs", tempDir}, command...)
	cmd := exec.Command("/proc/self/exe", execArgs...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = buildIsolationProcAttr()

	// Start the container process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the container process: %v", err)
	}

	// Attach the process to the cgroup
	if err := cgroup.AddProcess(cgroupPath, cmd.Process.Pid); err != nil {
		return fmt.Errorf("failed to attach the process to the cgroup: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait the container process: %v", err)
	}

	// Cleanup after container exits
	_ = os.RemoveAll(cgroupPath)

	return nil
}

func buildIsolationProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{

		Cloneflags: syscall.CLONE_NEWUTS | // Isolates hostname and domain name. Allows the container to have its own hostname.
			syscall.CLONE_NEWPID | // Creates a new PID namespace. The child becomes PID 1 inside the container.
			syscall.CLONE_NEWNS | // Creates a new mount namespace. Allows independent mount/unmount operations.
			syscall.CLONE_NEWCGROUP | // Isolates cgroup membership. Required for proper container cgroup hierarchy.
			syscall.CLONE_NEWUSER, // Creates a new user namespace. Allows “root inside container” without host root privileges.

		Unshareflags: syscall.CLONE_NEWNS, // Forces the current process to unshare the mount namespace before exec. Prevents mount propagation back to the host.

		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,           // root in container User namespace
				HostID:      os.Getuid(), // current user UID
				Size:        1,
			},
		},

		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,           // root in container User namespace
				HostID:      os.Getgid(), // current user GID
				Size:        1,
			},
		},
	}
}
