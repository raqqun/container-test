package main

import (
	"fmt"
	"os"

	"container-test-cli/internal/cli"
	"container-test-cli/internal/config"
	"container-test-cli/internal/output"
	"container-test-cli/internal/runner"
)

// runTests executes all test cases sequentially, respecting fail-fast behavior.
// Returns a slice of test results and the total number of failed tests.
func runTests(cfg *cli.CliConfig, tests []config.TestCase) ([]runner.Result, int) {
	enableColor := output.ShouldUseColor()
	results := make([]runner.Result, 0, len(tests))
	failures := 0

	for idx, testCase := range tests {
		name := testCase.ResolveName(idx)
		fmt.Printf("==> %s\n", name)

		res := runner.RunSingleTest(testCase, cfg.Engine, cfg.Image, cfg.DefaultTimeout, cfg.Debug, cfg.DryRun)

		if !cfg.DryRun {
			output.PrintResult(res, enableColor)
			if len(res.Failures) > 0 {
				failures++
			}
		}

		results = append(results, res)

		if cfg.FailFast && failures > 0 {
			fmt.Println("Stopping due to fail-fast")
			break
		}
	}

	return results, failures
}

// main is the entry point for the container test CLI. It parses command-line flags,
// loads test definitions, executes all tests, optionally writes a JSON report,
// and exits with an appropriate status code.
func main() {
	cfg := cli.ParseFlags()

	tests, err := config.LoadTests(cfg.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load tests: %v\n", err)
		os.Exit(2)
	}

	results, failures := runTests(cfg, tests)

	if cfg.JsonReport != "" {
		if err := output.WriteReport(cfg.JsonReport, results); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write report: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("Report written to %s\n", cfg.JsonReport)
	}

	if failures > 0 {
		fmt.Printf("\nCompleted with %d failing test(s)\n", failures)
		os.Exit(1)
	}
	fmt.Println("\nAll tests passed")
}
