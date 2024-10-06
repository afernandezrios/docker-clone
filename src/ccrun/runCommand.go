package ccrun

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var ccrunCmd = &cobra.Command{
	Use:   "ccrun",
	Short: "ccrun command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(ccrunCmd)
}

func runCommand(args []string) (e error) {

	cmd := exec.Command(args[0], args[1:]...)
	var serr bytes.Buffer
	cmd.Stderr = &serr
	stdout, err := cmd.Output()
	
	if err != nil {
		fmt.Printf(" %s\n", serr.String())
		return err
	}
	
	fmt.Printf("I got this output: %s\n", stdout)
	return nil
}
