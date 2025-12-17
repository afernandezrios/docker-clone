package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return runCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runCommand(args []string) (e error) {

	containerConfig := getContainerConfig()

	syscall.Sethostname([]byte("new-hostname"))

	chRootPath := "../container-dir"
	log.Printf("Root path: %s\n", chRootPath)
	syscall.Chroot(chRootPath)

	chDirPath := containerConfig.Config.WorkingDir
	log.Printf("Working dir: %s\n", chDirPath)
	syscall.Chdir(chDirPath)

	syscall.Mount("proc", "proc", "proc", 0, "")
	defer func() {
		syscall.Unmount("/proc", 0)
	}()

	// Export env variables. We must update Path variable
	// If not, probably it doesn't found any command because
	// it won't found /bin dirs in PATH
	for _, env := range containerConfig.Config.Env {
		log.Printf("New env variable: %s\n", env)
		envValues := strings.Split(env, "=")
		syscall.Setenv(envValues[0], envValues[1])
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Printf("Error running command. %s\n", err)
		return err
	}

	return nil
}

type ContainerConfig struct {
	Config ConfigData `json:"config"`
}

type ConfigData struct {
	Env        []string `json:"Env"`
	Entrypoint []string `json:"Entrypoint"`
	Volumes    []string `json:"Volumes"`
	WorkingDir string   `json:"WorkingDir"`
}

func getContainerConfig() ContainerConfig {
	byteValue, err := os.ReadFile("../container-dir/config.json")
	if err != nil {
		log.Panicf("Error reading configuration file: %v", err)
	}
	var containerConfig ContainerConfig
	json.Unmarshal(byteValue, &containerConfig)

	return containerConfig
}
