package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"container-test-cli/internal/config"
	"container-test-cli/internal/env"
	"container-test-cli/internal/output"
	"container-test-cli/internal/runner"
	"container-test-cli/internal/version"
)

// cliConfig holds all command-line configuration.
type cliConfig struct {
	configPath     string
	image          string
	engine         string
	defaultTimeout int
	jsonReport     string
	failFast       bool
	debug          bool
	dryRun         bool
	showVersion    bool
}

// parseFlags parses and validates command-line flags.
func parseFlags() *cliConfig {
	cfg := &cliConfig{}

	flag.StringVar(&cfg.configPath, "config", os.Getenv("CONTAINER_TEST_CONFIG"), "Path to YAML file describing tests")
	flag.StringVar(&cfg.image, "image", os.Getenv("CONTAINER_TEST_IMAGE"), "Image reference to run")
	flag.StringVar(&cfg.engine, "engine", env.EnvDefault("CONTAINER_TEST_ENGINE", "docker"), "Container engine CLI to use (docker, podman, ...)")
	flag.IntVar(&cfg.defaultTimeout, "default-timeout", env.EnvInt("CONTAINER_TEST_DEFAULT_TIMEOUT", 30), "Default timeout (seconds) for each test when not specified")
	flag.StringVar(&cfg.jsonReport, "json-report", os.Getenv("CONTAINER_TEST_JSON_REPORT"), "Write a JSON report to the given path")
	flag.BoolVar(&cfg.failFast, "fail-fast", env.EnvBool("CONTAINER_TEST_FAIL_FAST", false), "Stop on first failure")
	flag.BoolVar(&cfg.debug, "debug", env.EnvBool("CONTAINER_TEST_DEBUG", false), "Print commands before execution")
	flag.BoolVar(&cfg.dryRun, "dry-run", env.EnvBool("CONTAINER_TEST_DRY_RUN", false), "Print commands without executing")
	flag.BoolVar(&cfg.showVersion, "version", false, "Print version and exit")

	flag.Parse()

	if cfg.showVersion {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if cfg.configPath == "" || cfg.image == "" {
		fmt.Fprintln(os.Stderr, "config and image are required")
		flag.Usage()
		os.Exit(2)
	}

	return cfg
}

// testName returns the test name or generates a default one.
func testName(testCase config.TestCase, index int) string {
	if testCase.Name != "" {
		return testCase.Name
	}
	return fmt.Sprintf("test-%d", index+1)
}

// command resolves the command from either Exec or Command field.
func command(testCase config.TestCase) []string {
	if len(testCase.Exec) > 0 {
		return testCase.Exec
	}
	return testCase.Command
}

// handleDryRun prints the command that would be executed without running it.
func handleDryRun(engine, image string, testCase config.TestCase, name string) runner.Result {
	command := command(testCase)
	runCmd := runner.BuildRunCommand(engine, image, command, testCase.Workdir, testCase.Env, testCase.RunArgs, testCase.Entrypoint)
	fmt.Printf("   [dry-run] %s\n", strings.Join(runCmd, " "))
	return runner.Result{
		Status: "DRY-RUN",
		Name:   name,
	}
}

// printTestResult displays the test result with color and failure details.
func printTestResult(res runner.Result, enableColor bool) {
	statusColored := output.Colorize(res.Status, res.Status, enableColor)
	fmt.Printf("   %s\n", statusColored)
	for _, failure := range res.Failures {
		fmt.Printf("     - %s\n", failure)
	}
}

// runTests executes all tests and returns the results and failure count.
func runTests(cfg *cliConfig, tests []config.TestCase) ([]runner.Result, int) {
	enableColor := os.Getenv("NO_COLOR") == ""
	results := make([]runner.Result, 0, len(tests))
	failures := 0

	for idx, testCase := range tests {
		name := testName(testCase, idx)
		fmt.Printf("==> %s\n", name)

		var res runner.Result
		if cfg.dryRun {
			res = handleDryRun(cfg.engine, cfg.image, testCase, name)
		} else {
			res = runner.RunSingleTest(cfg.engine, cfg.image, testCase, cfg.defaultTimeout, cfg.debug)
			printTestResult(res, enableColor)
			if len(res.Failures) > 0 {
				failures++
			}
		}

		results = append(results, res)
		if cfg.failFast && failures > 0 {
			fmt.Println("Stopping due to fail-fast")
			break
		}
	}

	return results, failures
}

// main parses flags, loads tests, executes them, and sets the exit code.
func main() {
	cfg := parseFlags()

	tests, err := config.LoadTests(cfg.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tests: %v\n", err)
		os.Exit(2)
	}

	results, failures := runTests(cfg, tests)

	if cfg.jsonReport != "" {
		if err := output.WriteReport(cfg.jsonReport, results); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write report: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("Wrote report to %s\n", cfg.jsonReport)
	}

	if failures > 0 {
		fmt.Printf("Completed with %d failing test(s)\n", failures)
		os.Exit(1)
	}
	fmt.Println("All tests passed")
}
