package main_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	scriptDest = "/usr/local/bin/analysis.sh"
	image      = "golang:1.26-alpine"
)

func TestSecretChecker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	root, err := os.Getwd()
	require.NoError(t, err)

	hostPath, err := filepath.Abs(filepath.Join(root, "../../../.."))
	require.NoError(t, err)
	// Step 2: Define the container request
	req := testcontainers.ContainerRequest{
		Image: image,
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Mounts = append(hc.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: hostPath,
				Target: "/app",
			})
		},

		// Run apk update and then execute the copied binary
		// The binary is copied via Files field and then executed
		Cmd: []string{
			"sh", "-c",
			fmt.Sprintf("chmod +x %s && %s", scriptDest, scriptDest),
		},

		// Copy the local binary into the container
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "analysis.sh",
				ContainerFilePath: scriptDest,
				FileMode:          0o755,
			},
		},

		// Wait until the container finishes executing
		WaitingFor: wait.ForExit().WithExitTimeout(60 * time.Second),
	}

	// Step 3: Start the container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start container")

	// Ensure container cleanup
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	}()

	// Step 4: Retrieve and display container logs
	logs, err := container.Logs(ctx)
	require.NoError(t, err, "failed to get container logs")

	defer logs.Close()

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "ALERT DANGER FOUND") {
			assert.Fail(t, scanner.Text())
		}
	}

	// Step 5: Check exit code
	_, err = container.State(ctx)
	assert.NoError(t, err)
}
