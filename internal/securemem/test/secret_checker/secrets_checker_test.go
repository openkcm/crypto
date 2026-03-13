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

	isPersistentVaultGetFound := false
	isExposedSecretFound := false

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		assert.NotContains(t, scanner.Text(), "PANIC RECOVERED")
		assert.NotContains(t, scanner.Text(), "ALERT UNEXPOSED SECRET FOUND: MYSECRETKEY123458901234567890123")

		if !isPersistentVaultGetFound {
			isPersistentVaultGetFound = strings.Contains(scanner.Text(), "SECRET FOUND IN MEMVAULT")
		}
		if !isExposedSecretFound {
			isExposedSecretFound = strings.Contains(scanner.Text(), "EXPOSED SECRET FOUND: EXPOSED_SECRET123456789012345678")
		}
	}

	assert.True(t, isPersistentVaultGetFound, "persistent vault secret not found in logs")
	assert.True(t, isExposedSecretFound, "exposed secret not found in logs")

	// Step 5: Check exit code
	_, err = container.State(ctx)
	assert.NoError(t, err)
}
