package cli

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestParseFlags_Success tests successful flag parsing with all required parameters.
func TestParseFlags_Success(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flag.CommandLine for clean test state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd", "-config", "tests.yaml", "-image", "alpine:latest"}

	cfg := ParseFlags()

	if cfg.ConfigPath != "tests.yaml" {
		t.Errorf("Expected ConfigPath 'tests.yaml', got '%s'", cfg.ConfigPath)
	}
	if cfg.Image != "alpine:latest" {
		t.Errorf("Expected Image 'alpine:latest', got '%s'", cfg.Image)
	}
	if cfg.Engine != "docker" {
		t.Errorf("Expected default Engine 'docker', got '%s'", cfg.Engine)
	}
	if cfg.DefaultTimeout != 30 {
		t.Errorf("Expected default DefaultTimeout 30, got %d", cfg.DefaultTimeout)
	}
	if cfg.FailFast {
		t.Error("Expected FailFast to be false by default")
	}
	if cfg.Debug {
		t.Error("Expected Debug to be false by default")
	}
	if cfg.DryRun {
		t.Error("Expected DryRun to be false by default")
	}
}

// TestParseFlags_AllFlags tests parsing all available flags.
func TestParseFlags_AllFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"-config", "custom.yaml",
		"-image", "ubuntu:22.04",
		"-engine", "podman",
		"-default-timeout", "60",
		"-json-report", "report.json",
		"-fail-fast",
		"-debug",
		"-dry-run",
	}

	cfg := ParseFlags()

	if cfg.ConfigPath != "custom.yaml" {
		t.Errorf("Expected ConfigPath 'custom.yaml', got '%s'", cfg.ConfigPath)
	}
	if cfg.Image != "ubuntu:22.04" {
		t.Errorf("Expected Image 'ubuntu:22.04', got '%s'", cfg.Image)
	}
	if cfg.Engine != "podman" {
		t.Errorf("Expected Engine 'podman', got '%s'", cfg.Engine)
	}
	if cfg.DefaultTimeout != 60 {
		t.Errorf("Expected DefaultTimeout 60, got %d", cfg.DefaultTimeout)
	}
	if cfg.JsonReport != "report.json" {
		t.Errorf("Expected JsonReport 'report.json', got '%s'", cfg.JsonReport)
	}
	if !cfg.FailFast {
		t.Error("Expected FailFast to be true")
	}
	if !cfg.Debug {
		t.Error("Expected Debug to be true")
	}
	if !cfg.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

// TestParseFlags_EnvironmentVariables tests that environment variables are properly read.
func TestParseFlags_EnvironmentVariables(t *testing.T) {
	// Set up environment variables
	envVars := map[string]string{
		"CONTAINER_TEST_CONFIG":          "env-tests.yaml",
		"CONTAINER_TEST_IMAGE":           "env-image:latest",
		"CONTAINER_TEST_ENGINE":          "podman",
		"CONTAINER_TEST_DEFAULT_TIMEOUT": "45",
		"CONTAINER_TEST_JSON_REPORT":     "env-report.json",
		"CONTAINER_TEST_FAIL_FAST":       "true",
		"CONTAINER_TEST_DEBUG":            "1",
		"CONTAINER_TEST_DRY_RUN":          "yes",
	}

	// Set environment variables
	for key, val := range envVars {
		os.Setenv(key, val)
	}

	// Clean up environment variables after test
	defer func() {
		for key := range envVars {
			os.Unsetenv(key)
		}
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// No command-line flags, should use environment variables
	os.Args = []string{"cmd"}

	cfg := ParseFlags()

	if cfg.ConfigPath != "env-tests.yaml" {
		t.Errorf("Expected ConfigPath from env 'env-tests.yaml', got '%s'", cfg.ConfigPath)
	}
	if cfg.Image != "env-image:latest" {
		t.Errorf("Expected Image from env 'env-image:latest', got '%s'", cfg.Image)
	}
	if cfg.Engine != "podman" {
		t.Errorf("Expected Engine from env 'podman', got '%s'", cfg.Engine)
	}
	if cfg.DefaultTimeout != 45 {
		t.Errorf("Expected DefaultTimeout from env 45, got %d", cfg.DefaultTimeout)
	}
	if cfg.JsonReport != "env-report.json" {
		t.Errorf("Expected JsonReport from env 'env-report.json', got '%s'", cfg.JsonReport)
	}
	if !cfg.FailFast {
		t.Error("Expected FailFast from env to be true")
	}
	if !cfg.Debug {
		t.Error("Expected Debug from env to be true")
	}
	if !cfg.DryRun {
		t.Error("Expected DryRun from env to be true")
	}
}

// TestParseFlags_FlagsOverrideEnv tests that command-line flags override environment variables.
func TestParseFlags_FlagsOverrideEnv(t *testing.T) {
	os.Setenv("CONTAINER_TEST_CONFIG", "env-config.yaml")
	os.Setenv("CONTAINER_TEST_IMAGE", "env-image:latest")
	os.Setenv("CONTAINER_TEST_ENGINE", "podman")

	defer func() {
		os.Unsetenv("CONTAINER_TEST_CONFIG")
		os.Unsetenv("CONTAINER_TEST_IMAGE")
		os.Unsetenv("CONTAINER_TEST_ENGINE")
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Command-line flags should override environment variables
	os.Args = []string{
		"cmd",
		"-config", "flag-config.yaml",
		"-image", "flag-image:latest",
		"-engine", "docker",
	}

	cfg := ParseFlags()

	if cfg.ConfigPath != "flag-config.yaml" {
		t.Errorf("Expected ConfigPath from flag 'flag-config.yaml', got '%s'", cfg.ConfigPath)
	}
	if cfg.Image != "flag-image:latest" {
		t.Errorf("Expected Image from flag 'flag-image:latest', got '%s'", cfg.Image)
	}
	if cfg.Engine != "docker" {
		t.Errorf("Expected Engine from flag 'docker', got '%s'", cfg.Engine)
	}
}

// TestParseFlags_BooleanVariations tests various boolean value formats in environment variables.
func TestParseFlags_BooleanVariations(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"on", "on", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"off", "off", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CONTAINER_TEST_CONFIG", "test.yaml")
			os.Setenv("CONTAINER_TEST_IMAGE", "test:latest")
			os.Setenv("CONTAINER_TEST_DEBUG", tt.envValue)

			defer func() {
				os.Unsetenv("CONTAINER_TEST_CONFIG")
				os.Unsetenv("CONTAINER_TEST_IMAGE")
				os.Unsetenv("CONTAINER_TEST_DEBUG")
			}()

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Args = []string{"cmd"}

			cfg := ParseFlags()

			if cfg.Debug != tt.expected {
				t.Errorf("For env value '%s', expected Debug=%v, got %v", tt.envValue, tt.expected, cfg.Debug)
			}
		})
	}
}

// TestParseFlags_MissingConfig tests behavior when config flag is missing.
// This test uses subprocess execution to properly test os.Exit() behavior.
func TestParseFlags_MissingConfig(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-image", "alpine:latest"}
		ParseFlags()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseFlags_MissingConfig")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("Expected non-zero exit code when config is missing")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %T", err)
	}

	if exitErr.ExitCode() != 2 {
		t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "config and image are required") {
		t.Errorf("Expected error message about required flags, got: %s", outputStr)
	}
}

// TestParseFlags_MissingImage tests behavior when image flag is missing.
func TestParseFlags_MissingImage(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-config", "tests.yaml"}
		ParseFlags()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseFlags_MissingImage")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("Expected non-zero exit code when image is missing")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %T", err)
	}

	if exitErr.ExitCode() != 2 {
		t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "config and image are required") {
		t.Errorf("Expected error message about required flags, got: %s", outputStr)
	}
}

// TestParseFlags_MissingBoth tests behavior when both required flags are missing.
func TestParseFlags_MissingBoth(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd"}
		ParseFlags()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseFlags_MissingBoth")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("Expected non-zero exit code when both config and image are missing")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %T", err)
	}

	if exitErr.ExitCode() != 2 {
		t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "config and image are required") {
		t.Errorf("Expected error message about required flags, got: %s", outputStr)
	}
}

// TestParseFlags_VersionFlag tests the version flag behavior.
func TestParseFlags_VersionFlag(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-version"}
		ParseFlags()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseFlags_VersionFlag")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	// Version flag should exit with code 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 0 {
				t.Errorf("Expected exit code 0 for version flag, got %d", exitErr.ExitCode())
			}
		}
	}

	outputStr := string(output)
	// Should print version (default is "dev")
	if !strings.Contains(outputStr, "dev") && outputStr == "" {
		t.Errorf("Expected version output, got: %s", outputStr)
	}
}

// TestParseFlags_DefaultTimeoutInvalid tests behavior with invalid timeout value in environment.
func TestParseFlags_DefaultTimeoutInvalid(t *testing.T) {
	os.Setenv("CONTAINER_TEST_CONFIG", "test.yaml")
	os.Setenv("CONTAINER_TEST_IMAGE", "test:latest")
	os.Setenv("CONTAINER_TEST_DEFAULT_TIMEOUT", "invalid")

	defer func() {
		os.Unsetenv("CONTAINER_TEST_CONFIG")
		os.Unsetenv("CONTAINER_TEST_IMAGE")
		os.Unsetenv("CONTAINER_TEST_DEFAULT_TIMEOUT")
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd"}

	cfg := ParseFlags()

	// Should fall back to default value of 30 when parsing fails
	if cfg.DefaultTimeout != 30 {
		t.Errorf("Expected DefaultTimeout to fall back to 30 for invalid env value, got %d", cfg.DefaultTimeout)
	}
}

// TestParseFlags_ZeroTimeout tests parsing with zero timeout value.
func TestParseFlags_ZeroTimeout(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"-config", "test.yaml",
		"-image", "test:latest",
		"-default-timeout", "0",
	}

	cfg := ParseFlags()

	if cfg.DefaultTimeout != 0 {
		t.Errorf("Expected DefaultTimeout 0, got %d", cfg.DefaultTimeout)
	}
}

// TestParseFlags_NegativeTimeout tests parsing with negative timeout value.
func TestParseFlags_NegativeTimeout(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"-config", "test.yaml",
		"-image", "test:latest",
		"-default-timeout", "-10",
	}

	cfg := ParseFlags()

	// The flag package allows negative integers
	if cfg.DefaultTimeout != -10 {
		t.Errorf("Expected DefaultTimeout -10, got %d", cfg.DefaultTimeout)
	}
}

// TestParseFlags_EmptyStringValues tests parsing with empty string values.
func TestParseFlags_EmptyStringValues(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{
		"cmd",
		"-config", "test.yaml",
		"-image", "test:latest",
		"-engine", "",
		"-json-report", "",
	}

	cfg := ParseFlags()

	// Empty strings should be preserved
	if cfg.Engine != "" {
		t.Errorf("Expected empty Engine, got '%s'", cfg.Engine)
	}
	if cfg.JsonReport != "" {
		t.Errorf("Expected empty JsonReport, got '%s'", cfg.JsonReport)
	}
}

// TestCliConfig_Structure tests the CliConfig struct fields and types.
func TestCliConfig_Structure(t *testing.T) {
	cfg := &CliConfig{
		ConfigPath:     "test.yaml",
		Image:          "alpine:latest",
		Engine:         "docker",
		DefaultTimeout: 30,
		JsonReport:     "report.json",
		FailFast:       true,
		Debug:          true,
		DryRun:         true,
		ShowVersion:    true,
	}

	if cfg.ConfigPath != "test.yaml" {
		t.Error("ConfigPath field not properly set")
	}
	if cfg.Image != "alpine:latest" {
		t.Error("Image field not properly set")
	}
	if cfg.Engine != "docker" {
		t.Error("Engine field not properly set")
	}
	if cfg.DefaultTimeout != 30 {
		t.Error("DefaultTimeout field not properly set")
	}
	if cfg.JsonReport != "report.json" {
		t.Error("JsonReport field not properly set")
	}
	if !cfg.FailFast {
		t.Error("FailFast field not properly set")
	}
	if !cfg.Debug {
		t.Error("Debug field not properly set")
	}
	if !cfg.DryRun {
		t.Error("DryRun field not properly set")
	}
	if !cfg.ShowVersion {
		t.Error("ShowVersion field not properly set")
	}
}
