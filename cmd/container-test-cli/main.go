package main

import (
	"fmt"
	"os"

	"container-test-cli/internal/cli"
	"container-test-cli/internal/config"
	"container-test-cli/internal/output"
	"container-test-cli/internal/runner"
)

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

	results, failures := runner.RunTests(tests, runner.Config{
		Engine:         cfg.Engine,
		Image:          cfg.Image,
		DefaultTimeout: cfg.DefaultTimeout,
		FailFast:       cfg.FailFast,
		Debug:          cfg.Debug,
		DryRun:         cfg.DryRun,
	})

	// Print results
	if !cfg.DryRun {
		enableColor := output.ShouldUseColor()
		for _, res := range results {
			output.PrintResult(res, enableColor)
		}
	}

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
