package main

import (
	"flag"
	"fmt"
	"os"
)

var (
    version   string
    commitSHA string
)

// main parses flags, loads tests, executes them, and sets the exit code.
func main() {
    var (
        configPath     string
        image          string
        engine         string
        defaultTimeout int
        jsonReport     string
        failFast       bool
        showVersion    bool
        dryRun         bool
        debug          bool
    )

    flag.StringVar(&configPath, "config", os.Getenv("CONTAINER_TEST_CONFIG"), "Path to YAML file describing tests")
    flag.StringVar(&image, "image", os.Getenv("CONTAINER_TEST_IMAGE"), "Image reference to run")
    flag.StringVar(&engine, "engine", getenvDefault("CONTAINER_TEST_ENGINE", "docker"), "Container engine CLI to use (docker, podman, ...)")
    flag.IntVar(&defaultTimeout, "default-timeout", envInt("CONTAINER_TEST_DEFAULT_TIMEOUT", 30), "Default timeout (seconds) for each test when not specified")
    flag.StringVar(&jsonReport, "json-report", os.Getenv("CONTAINER_TEST_JSON_REPORT"), "Write a JSON report to the given path")
    flag.BoolVar(&failFast, "fail-fast", envBool("CONTAINER_TEST_FAIL_FAST", false), "Stop on first failure")
    flag.BoolVar(&showVersion, "version", false, "Print version and exit")
    flag.BoolVar(&dryRun, "dry-run", envBool("CONTAINER_TEST_DRY_RUN", false), "Print commands without executing them")
    flag.BoolVar(&debug, "debug", envBool("CONTAINER_TEST_DEBUG", false), "Print commands before execution")

    flag.Parse()

    if showVersion {
        fmt.Printf(fmt.Sprintf("%s-%s\n", version, commitSHA))
        os.Exit(0)
    }

    if configPath == "" || image == "" {
        fmt.Fprintln(os.Stderr, "config and image are required")
        flag.Usage()
        os.Exit(2)
    }

    tests, err := loadTests(configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load tests: %v\n", err)
        os.Exit(2)
    }

    enableColor := os.Getenv("NO_COLOR") == ""
    results := make([]testResult, 0, len(tests))
    failures := 0

    for idx, t := range tests {
        name := t.Name
        if name == "" {
            name = fmt.Sprintf("test-%d", idx+1)
        }
        fmt.Printf("==> %s\n", name)
        res := runSingleTest(engine, image, t, defaultTimeout, debug, dryRun)
        statusColored := colorize(res.Status, res.Status, enableColor)
        fmt.Printf("   %s\n", statusColored)
        if len(res.Failures) > 0 {
            failures++
            for _, f := range res.Failures {
                fmt.Printf("     - %s\n", f)
            }
        }
        results = append(results, res)
        if failFast && failures > 0 {
            fmt.Println("Stopping due to fail-fast")
            break
        }
    }

    if jsonReport != "" {
        if err := writeReport(jsonReport, results); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to write report: %v\n", err)
            os.Exit(2)
        }
        fmt.Printf("Wrote report to %s\n", jsonReport)
    }

    if failures > 0 {
        fmt.Printf("Completed with %d failing test(s)\n", failures)
        os.Exit(1)
    }
    fmt.Println("All tests passed")
}
