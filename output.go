package main

import (
	"encoding/json"
	"os"
)

const (
    green  = "\033[32m"
    red    = "\033[31m"
    yellow = "\033[33m"
    reset  = "\033[0m"
)

// colorize applies ANSI color to status text when enabled.
func colorize(text, status string, enable bool) string {
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

// writeReport emits the JSON results file if requested.
func writeReport(path string, results []testResult) error {
    data, err := json.MarshalIndent(results, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0o644)
}
