package ccrun

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

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

	cmd := exec.Command("/proc/self/exe", append([]string{"run"}, args[0:]...)...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
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
		fmt.Printf(" %s\n", err)
		return err
	}

	return nil
}
