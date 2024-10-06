package ccrun

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var ccrunCmd = &cobra.Command{
	Use:   "ccrun",
	Short: "ccrun command",
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
	// 1. Create new hostname: sudo unshare --uts /bin/bash && hostname new_host
	// 2. Run command inside the new hostname

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf(" %s\n", err)
		return err
	}

	return nil
}
