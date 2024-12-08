package ccrun

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/afernandezrios/docker-clone/ccrun/docker"
	"github.com/spf13/cobra"
)

var runInNamespaceCmd = &cobra.Command{
	Use:   "ccrun",
	Short: "ccrun command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return runInNewUTSNamespace(args)
	},
}

func init() {
	rootCmd.AddCommand(runInNamespaceCmd)
}

func runInNewUTSNamespace(args []string) (e error) {

	docker.DownloadImage()

	// TODO: delete cgroup on exit?
	cgroup := NewCgroup()
	log.Printf("New cgroup created: %s\n", cgroup)

	cmd := exec.Command("/proc/self/exe", append([]string{"run"}, args[0:]...)...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWCGROUP |
			syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
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

	if err := cmd.Run(); err != nil {
		log.Printf(" %s\n", err)
		return err
	}

	return nil
}
