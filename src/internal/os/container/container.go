package container

import (
	"strings"
	"syscall"

	"github.com/afernandezrios/docker-clone/config"
)

func Setup(imgConfig *config.ImageConfig, rootfs string) error {
	syscall.Sethostname([]byte("new-hostname"))

	// Change root filesystem
	if err := syscall.Chroot(rootfs); err != nil {
		return err
	}

	// Set working directory inside the new root
	if err := syscall.Chdir(imgConfig.WorkingDir); err != nil {
		return err
	}

	// Apply environment variables
	for _, env := range imgConfig.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		syscall.Setenv(parts[0], parts[1])
	}
	return nil
}
