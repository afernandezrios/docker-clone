package ccrun

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	containerConfig := getContainerConfig()

	syscall.Sethostname([]byte("new-hostname"))
	syscall.Chroot("../container-dir")
	syscall.Chdir(containerConfig.Config.WorkingDir)
	syscall.Mount("proc", "proc", "proc", 0, "")
	
	// Export env variables. We must update Path variable
	// If not, probably it doesn't found any command because
	// it won't found /bin dirs in PATH
	for _, env := range containerConfig.Config.Env {
		envValues := strings.Split(env, "=")
		syscall.Setenv(envValues[0], envValues[1])
	}
	printEnvVar("PATH")

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

func printEnvVar(varName string) {
	value, found := syscall.Getenv(varName)
    if !found {
        fmt.Println("Environment variable not found")
        return
    }

    fmt.Println("PATH:", value)
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
	byteValue, _ := os.ReadFile("../container-dir/config.json")
	var containerConfig ContainerConfig
	json.Unmarshal(byteValue, &containerConfig)

	return containerConfig
}
