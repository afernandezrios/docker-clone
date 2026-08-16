package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_RunContainer(t *testing.T) {
	// 1. Skip if not running on Linux or without root privileges
	if os.Getuid() != 0 {
		t.Skip("Skipping integration test: root privileges (sudo) required for namespaces/cgroups")
	}

	// 2. Build the binary to a temporary directory
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "ccrun")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build ccrun binary: %v\nOutput: %s", err, string(out))
	}

	// 3. Define the test case
	expectedMessage := "container-runtime-integration-test-ok"
	
	// Command: ./docker-clone ccrun library/alpine /bin/echo <expectedMessage>
	cmd := exec.Command(binaryPath, "ccrun", "library/alpine", "/bin/echo", expectedMessage)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 4. Run the container command
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Container execution failed: %v\nStderr: %s\nStdout: %s", err, stderr.String(), stdout.String())
	}

	// 5. Assert output
	actualOutput := strings.TrimSpace(stdout.String())
	if !strings.Contains(actualOutput, expectedMessage) {
		t.Errorf("Expected output to contain %q, but got %q", expectedMessage, actualOutput)
	}
}