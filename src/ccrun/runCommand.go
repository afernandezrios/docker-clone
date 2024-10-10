package ccrun

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var ccrunCmd = &cobra.Command{
	Use:   "run",
	Short: "run command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return runCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(ccrunCmd)
}

func runCommand(args []string) (e error) {

	syscall.Sethostname([]byte("new-hostname"))
	syscall.Chroot("../alpine")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf(" %s\n", err)
		syscall.Unmount("/proc", 0)
		return err
	}

	syscall.Unmount("/proc", 0)
	return nil
}
