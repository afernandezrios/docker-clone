package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/afernandezrios/docker-clone/config"
	"github.com/afernandezrios/docker-clone/internal/os/virtual_filesystems"
	"github.com/afernandezrios/docker-clone/internal/os/container"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runContainerProcess(args)
	},
}

func runContainerProcess(args []string) (e error) {

	containerConfig := config.GetContainer()

	// Isolate the environment
	if err := container.Setup(containerConfig); err != nil {
		log.Printf("failed to setup environment: %s", err)
	}

	// Mount and cleanup virtual filesystems
	cleanup, err := virtual_filesystems.Mount()
	if err != nil {
		log.Printf("failed to mount filesystems: %s", err)
	}
	defer cleanup()

	// Execute the target command
	return executeProcess(args)
}

func executeProcess(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Printf("Process execution failed: %v", err)
		return err
	}
	return nil
}
