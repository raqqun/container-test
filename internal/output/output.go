package output

import (
	"encoding/json"
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
