package output

import (
	"encoding/json"
	"fmt"
	"os"

	"container-test-cli/internal/runner"
)

const (
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

// Colorize applies ANSI color to status text when enabled.
func Colorize(text, status string, enable bool) string {
	if !enable {
		return text
	}
	switch status {
	case "PASSED":
		return green + text + reset
	case "FAILED":
		return red + text + reset
	default:
		return yellow + text + reset
	}
}

// WriteReport emits the JSON results file if requested.
func WriteReport(path string, results []runner.Result) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// PrintResult displays the test result status with color formatting.
func PrintResult(res runner.Result, enableColor bool) {
	statusColored := Colorize(res.Status, res.Status, enableColor)
	fmt.Printf("   %s\n", statusColored)
	for _, failure := range res.Failures {
		fmt.Printf("     - %s\n", failure)
	}
}

// ShouldUseColor returns true if color output should be enabled.
func ShouldUseColor() bool {
	return os.Getenv("NO_COLOR") == ""
}
