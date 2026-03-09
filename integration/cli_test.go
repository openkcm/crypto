package integration

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openkcm/krypton/pkg/authn"
)

// binaryPath holds the path to the compiled kr binary.
var binaryPath string

// coverDir holds the path to the coverage directory (from GOCOVERDIR env var).
var coverDir string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "kr-integration-test-*")
	if err != nil {
		os.Exit(1)
	}
	binaryPath = filepath.Join(tmpDir, "kr")

	// Check if we should build with coverage instrumentation
	coverDir = os.Getenv("GOCOVERDIR")
	buildArgs := []string{"build", "-o", binaryPath}
	if coverDir != "" {
		// Use atomic mode to be compatible with -race flag
		buildArgs = append(buildArgs, "-cover", "-covermode=atomic")
	}
	buildArgs = append(buildArgs, "../cli")

	buildCmd := exec.CommandContext(context.Background(), "go", buildArgs...)
	buildCmd.Stderr = os.Stderr
	err = buildCmd.Run()
	if err != nil {
		os.Exit(1)
	}

	exitCode := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(exitCode)
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name        string
		setupToken  bool
		expContains string
	}{
		{
			name:        "login without existing token",
			setupToken:  false,
			expContains: "Login successful.",
		},
		{
			name:        "login with existing token",
			setupToken:  true,
			expContains: "Already logged in.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpHome := t.TempDir()

			if tt.setupToken {
				createTestToken(t, tmpHome)
			}

			cmd := newCommand(t.Context(), tmpHome, "login")

			output, err := cmd.CombinedOutput()
			assert.NoError(t, err)
			assert.Contains(t, string(output), tt.expContains)
		})
	}
}

func TestLogin_CreateTokenFile(t *testing.T) {
	tmpHome := t.TempDir()

	cmd := newCommand(t.Context(), tmpHome, "login")

	_, err := cmd.CombinedOutput()
	assert.NoError(t, err)

	tokenPath := filepath.Join(tmpHome, ".krypton", "token.json")
	_, err = os.Stat(tokenPath)
	assert.NoError(t, err, "token file should exist after login")

	data, err := os.ReadFile(tokenPath)
	assert.NoError(t, err)

	var token authn.Token
	err = json.Unmarshal(data, &token)
	assert.NoError(t, err)
}

func TestLogin_Help(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expContains string
	}{
		{
			name:        "login help with long flag",
			args:        []string{"login", "--help"},
			expContains: "Authenticate with the Krypton server",
		},
		{
			name:        "login help with short flag",
			args:        []string{"login", "-h"},
			expContains: "Authenticate with the Krypton server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(t.Context(), t.TempDir(), tt.args...)
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err)
			assert.Contains(t, string(output), tt.expContains)
		})
	}
}

// newCommand creates a new exec.Command with the given arguments and sets up
// the environment variables including HOME and GOCOVERDIR (if coverage is enabled).
func newCommand(ctx context.Context, homeDir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Env = []string{"HOME=" + homeDir}
	if coverDir != "" {
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+coverDir)
	}
	return cmd
}

func createTestToken(t *testing.T, homeDir string) {
	t.Helper()

	kryptonDir := filepath.Join(homeDir, ".krypton")
	err := os.MkdirAll(kryptonDir, 0700)
	assert.NoError(t, err)

	token := authn.Token{
		Type:      "bearer",
		Value:     []byte("test-token-value"),
		ExpiredAt: 9999999999,
		Attributes: map[string]any{
			"test": true,
		},
	}

	data, err := json.Marshal(token)
	assert.NoError(t, err)

	tokenPath := filepath.Join(kryptonDir, "token.json")
	err = os.WriteFile(tokenPath, data, 0600)
	assert.NoError(t, err)
}
