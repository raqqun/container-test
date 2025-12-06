package cli

import (
	"container-test-cli/internal/env"
	"container-test-cli/internal/version"
	"flag"
	"fmt"
	"os"
)

// CliConfig holds the parsed command-line configuration for the application.
type CliConfig struct {
	ConfigPath     string
	Image          string
	Engine         string
	DefaultTimeout int
	JsonReport     string
	FailFast       bool
	Debug          bool
	DryRun         bool
	ShowVersion    bool
}

// parseFlags parses command-line flags, validates required parameters, and returns the configuration.
// Exits with code 0 if --version is specified, or code 2 if validation fails.
func ParseFlags() *CliConfig {
	cfg := &CliConfig{}

	flag.StringVar(&cfg.ConfigPath, "config", os.Getenv("CONTAINER_TEST_CONFIG"), "Path to YAML file describing tests")
	flag.StringVar(&cfg.Image, "image", os.Getenv("CONTAINER_TEST_IMAGE"), "Image reference to run")
	flag.StringVar(&cfg.Engine, "engine", env.EnvDefault("CONTAINER_TEST_ENGINE", "docker"), "Container engine CLI to use (docker, podman, ...)")
	flag.IntVar(&cfg.DefaultTimeout, "default-timeout", env.EnvInt("CONTAINER_TEST_DEFAULT_TIMEOUT", 30), "Default timeout (seconds) for each test when not specified")
	flag.StringVar(&cfg.JsonReport, "json-report", os.Getenv("CONTAINER_TEST_JSON_REPORT"), "Write a JSON report to the given path")
	flag.BoolVar(&cfg.FailFast, "fail-fast", env.EnvBool("CONTAINER_TEST_FAIL_FAST", false), "Stop on first failure")
	flag.BoolVar(&cfg.Debug, "debug", env.EnvBool("CONTAINER_TEST_DEBUG", false), "Print commands before execution")
	flag.BoolVar(&cfg.DryRun, "dry-run", env.EnvBool("CONTAINER_TEST_DRY_RUN", false), "Print commands without executing")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Print version and exit")

	flag.Parse()

	if cfg.ShowVersion {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if cfg.ConfigPath == "" || cfg.Image == "" {
		fmt.Fprintln(os.Stderr, "Error: config and image are required")
		flag.Usage()
		os.Exit(2)
	}

	return cfg
}
