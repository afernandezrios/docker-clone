package ccrun

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	cmd := exec.Command("/bin/bash", "-c", "hostname new-hostname;" + strings.Join(args, ";"))
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
