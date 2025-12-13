package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"container-test-cli/internal/runner"
)

// TestColorize_Enabled tests colorization when enabled.
func TestColorize_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		status   string
		expected string
	}{
		{"passed", "test passed", "PASSED", green + "test passed" + reset},
		{"failed", "test failed", "FAILED", red + "test failed" + reset},
		{"skipped", "test skipped", "SKIPPED", yellow + "test skipped" + reset},
		{"dry-run", "test dry-run", "DRY-RUN", yellow + "test dry-run" + reset},
		{"unknown", "test unknown", "UNKNOWN", yellow + "test unknown" + reset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.text, tt.status, true)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestColorize_Disabled tests colorization when disabled.
func TestColorize_Disabled(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		status string
	}{
		{"passed", "test passed", "PASSED"},
		{"failed", "test failed", "FAILED"},
		{"skipped", "test skipped", "SKIPPED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.text, tt.status, false)
			if result != tt.text {
				t.Errorf("Expected plain text '%s', got '%s'", tt.text, result)
			}
		})
	}
}

// TestColorize_EmptyText tests colorization with empty text.
func TestColorize_EmptyText(t *testing.T) {
	result := Colorize("", "PASSED", true)
	expected := green + "" + reset
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestColorize_SpecialCharacters tests colorization with special characters.
func TestColorize_SpecialCharacters(t *testing.T) {
	text := "test\nwith\nnewlines\tand\ttabs"
	result := Colorize(text, "PASSED", true)
	expected := green + text + reset
	if result != expected {
		t.Errorf("Expected colorized text with special chars, got '%s'", result)
	}
}

// TestShouldUseColor_NoColorNotSet tests color detection when NO_COLOR is not set.
func TestShouldUseColor_NoColorNotSet(t *testing.T) {
	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	result := shouldUseColor()
	if !result {
		t.Error("Expected color to be enabled when NO_COLOR is not set")
	}
}

// TestShouldUseColor_NoColorSet tests color detection when NO_COLOR is set.
func TestShouldUseColor_NoColorSet(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	result := shouldUseColor()
	if result {
		t.Error("Expected color to be disabled when NO_COLOR is set")
	}
}

// TestShouldUseColor_NoColorEmpty tests color detection when NO_COLOR is empty.
func TestShouldUseColor_NoColorEmpty(t *testing.T) {
	os.Setenv("NO_COLOR", "")
	defer os.Unsetenv("NO_COLOR")

	result := shouldUseColor()
	if !result {
		t.Error("Expected color to be enabled when NO_COLOR is empty string")
	}
}

// TestWriteReport_ValidResults tests writing a valid JSON report.
func TestWriteReport_ValidResults(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.json")

	exitCode := 0
	results := []runner.Result{
		{
			Status:   "PASSED",
			Name:     "test-1",
			Stdout:   "hello",
			Stderr:   "",
			ExitCode: &exitCode,
		},
		{
			Status:   "FAILED",
			Name:     "test-2",
			Stdout:   "",
			Stderr:   "error",
			ExitCode: &exitCode,
			Failures: []string{"exit code mismatch"},
		},
	}

	err := WriteReport(reportPath, results)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var decoded []runner.Result
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded) != 2 {
		t.Errorf("Expected 2 results, got %d", len(decoded))
	}

	if decoded[0].Name != "test-1" {
		t.Errorf("Expected first test name 'test-1', got '%s'", decoded[0].Name)
	}

	if decoded[1].Status != "FAILED" {
		t.Errorf("Expected second test status 'FAILED', got '%s'", decoded[1].Status)
	}
}

// TestWriteReport_EmptyResults tests writing an empty results array.
func TestWriteReport_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "empty-report.json")

	results := []runner.Result{}

	err := WriteReport(reportPath, results)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var decoded []runner.Result
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("Expected 0 results, got %d", len(decoded))
	}
}

// TestWriteReport_NilResults tests writing nil results.
func TestWriteReport_NilResults(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "nil-report.json")

	err := WriteReport(reportPath, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	// nil slice marshals to "null" in JSON
	if string(data) != "null" {
		t.Errorf("Expected 'null', got '%s'", string(data))
	}
}

// TestWriteReport_InvalidPath tests writing to an invalid path.
func TestWriteReport_InvalidPath(t *testing.T) {
	invalidPath := "/nonexistent/directory/report.json"

	exitCode := 0
	results := []runner.Result{
		{
			Status:   "PASSED",
			Name:     "test",
			ExitCode: &exitCode,
		},
	}

	err := WriteReport(invalidPath, results)
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}
}

// TestWriteReport_ComplexResults tests writing results with all fields populated.
func TestWriteReport_ComplexResults(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "complex-report.json")

	exitCode := 1
	results := []runner.Result{
		{
			Status:        "FAILED",
			Name:          "complex-test",
			Stdout:        "output line 1\noutput line 2",
			Stderr:        "error line 1\nerror line 2",
			ExitCode:      &exitCode,
			Failures:      []string{"failure 1", "failure 2", "failure 3"},
			DebugMessages: []string{"debug msg 1", "debug msg 2"},
		},
	}

	err := WriteReport(reportPath, results)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var decoded []runner.Result
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded[0].Failures) != 3 {
		t.Errorf("Expected 3 failures, got %d", len(decoded[0].Failures))
	}

	if len(decoded[0].DebugMessages) != 2 {
		t.Errorf("Expected 2 debug messages, got %d", len(decoded[0].DebugMessages))
	}
}

// TestPrintResult_Passed tests printing a passed result.
func TestPrintResult_Passed(t *testing.T) {
	// Disable color for predictable output
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 0
	result := runner.Result{
		Status:   "PASSED",
		Name:     "test-passed",
		ExitCode: &exitCode,
	}

	PrintResult(result)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test-passed") {
		t.Errorf("Expected output to contain 'test-passed', got: %s", output)
	}

	if !strings.Contains(output, "PASSED") {
		t.Errorf("Expected output to contain 'PASSED', got: %s", output)
	}
}

// TestPrintResult_Failed tests printing a failed result with failures.
func TestPrintResult_Failed(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 1
	result := runner.Result{
		Status:   "FAILED",
		Name:     "test-failed",
		ExitCode: &exitCode,
		Failures: []string{"failure reason 1", "failure reason 2"},
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test-failed") {
		t.Errorf("Expected output to contain 'test-failed', got: %s", output)
	}

	if !strings.Contains(output, "FAILED") {
		t.Errorf("Expected output to contain 'FAILED', got: %s", output)
	}

	if !strings.Contains(output, "failure reason 1") {
		t.Errorf("Expected output to contain 'failure reason 1', got: %s", output)
	}

	if !strings.Contains(output, "failure reason 2") {
		t.Errorf("Expected output to contain 'failure reason 2', got: %s", output)
	}
}

// TestPrintResult_WithDebugMessages tests printing a result with debug messages.
func TestPrintResult_WithDebugMessages(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 0
	result := runner.Result{
		Status:        "PASSED",
		Name:          "test-debug",
		ExitCode:      &exitCode,
		DebugMessages: []string{"debug message 1", "debug message 2"},
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[debug] debug message 1") {
		t.Errorf("Expected output to contain debug message 1, got: %s", output)
	}

	if !strings.Contains(output, "[debug] debug message 2") {
		t.Errorf("Expected output to contain debug message 2, got: %s", output)
	}
}

// TestPrintResult_Skipped tests printing a skipped result.
func TestPrintResult_Skipped(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runner.Result{
		Status: "SKIPPED",
		Name:   "test-skipped",
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test-skipped") {
		t.Errorf("Expected output to contain 'test-skipped', got: %s", output)
	}

	if !strings.Contains(output, "SKIPPED") {
		t.Errorf("Expected output to contain 'SKIPPED', got: %s", output)
	}
}

// TestPrintResult_DryRun tests printing a dry-run result.
func TestPrintResult_DryRun(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runner.Result{
		Status: "DRY-RUN",
		Name:   "test-dryrun",
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test-dryrun") {
		t.Errorf("Expected output to contain 'test-dryrun', got: %s", output)
	}

	if !strings.Contains(output, "DRY-RUN") {
		t.Errorf("Expected output to contain 'DRY-RUN', got: %s", output)
	}
}

// TestPrintResult_WithColor tests that color codes are applied when enabled.
func TestPrintResult_WithColor(t *testing.T) {
	// Ensure NO_COLOR is not set
	os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 0
	result := runner.Result{
		Status:   "PASSED",
		Name:     "test-color",
		ExitCode: &exitCode,
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("Expected output to contain ANSI color codes, got: %s", output)
	}
}

// TestPrintResult_EmptyFailures tests printing with empty failures array.
func TestPrintResult_EmptyFailures(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 0
	result := runner.Result{
		Status:   "PASSED",
		Name:     "test-empty-failures",
		ExitCode: &exitCode,
		Failures: []string{},
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not have any failure markers
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "     - ") {
			t.Errorf("Expected no failure lines, but found: %s", line)
		}
	}
}

// TestPrintResult_EmptyDebugMessages tests printing with empty debug messages array.
func TestPrintResult_EmptyDebugMessages(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 0
	result := runner.Result{
		Status:        "PASSED",
		Name:          "test-empty-debug",
		ExitCode:      &exitCode,
		DebugMessages: []string{},
	}

	PrintResult(result)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not have any debug message markers
	if strings.Contains(output, "[debug]") {
		t.Errorf("Expected no debug messages, but found some in: %s", output)
	}
}

// TestWriteReport_FilePermissions tests that created file has correct permissions.
func TestWriteReport_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "perms-report.json")

	exitCode := 0
	results := []runner.Result{
		{
			Status:   "PASSED",
			Name:     "test",
			ExitCode: &exitCode,
		},
	}

	err := WriteReport(reportPath, results)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	info, err := os.Stat(reportPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Check file permissions (0644)
	mode := info.Mode().Perm()
	expected := os.FileMode(0644)
	if mode != expected {
		t.Errorf("Expected file permissions %o, got %o", expected, mode)
	}
}

// TestColorize_CaseSensitivity tests that status matching is case-sensitive.
func TestColorize_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"uppercase_passed", "PASSED", green + "text" + reset},
		{"lowercase_passed", "passed", yellow + "text" + reset}, // Should be yellow (default)
		{"uppercase_failed", "FAILED", red + "text" + reset},
		{"lowercase_failed", "failed", yellow + "text" + reset}, // Should be yellow (default)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize("text", tt.status, true)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
