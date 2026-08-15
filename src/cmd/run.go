package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/afernandezrios/docker-clone/config"
	"github.com/afernandezrios/docker-clone/internal/os/container"
	"github.com/afernandezrios/docker-clone/internal/os/vfs"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cleanup, err := setupEnvironment()
		if err != nil {
			return fmt.Errorf("setup environment: %w", err)
		}
		defer cleanup()

		if err := executeProcess(args); err != nil {
			return fmt.Errorf("Process execution failed: %w", err)
		}

		return nil
	},
}

func setupEnvironment() (func(), error) {

	containerConfig := config.GetContainer()

	// Isolate the environment
	if err := container.Setup(containerConfig); err != nil {
		return nil, fmt.Errorf("failed to setup environment: %w", err)
	}

	// Mount and cleanup virtual filesystems
	cleanup, err := vfs.Mount()
	if err != nil {
		return nil, fmt.Errorf("failed to mount filesystems: %w", err)
	}

	return cleanup, nil
}

func executeProcess(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
