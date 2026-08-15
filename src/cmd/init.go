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

var rootfsDir string

var initCmd = &cobra.Command{
	Use:    "init [command] [args...]",
	Short:  "Internal container initializer",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if rootfsDir == "" {
			return fmt.Errorf("--rootfs is required")
		}

		cleanup, err := setupEnvironment(rootfsDir)
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

func init() {
	initCmd.Flags().StringVar(&rootfsDir, "rootfs", "", "Path to container root directory")
	// Hide flag from public --help output if 'run' is internal to re-exec
	_ = initCmd.Flags().MarkHidden("rootfs")
}

func setupEnvironment(rootfs string) (func(), error) {

	imgConfig, err := config.Load(rootfs)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Isolate the environment
	if err := container.Setup(imgConfig, rootfs); err != nil {
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
