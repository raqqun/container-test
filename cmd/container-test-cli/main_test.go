package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"container-test-cli/internal/runner"
)

// TestMain_SuccessfulExecution tests the main function with all tests passing.
func TestMain_SuccessfulExecution(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		// Create temporary test config with passing tests
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_success.yaml")
		configContent := `tests:
  - name: test-success
    command: "echo hello"
    expect:
      exit_code: 0
      stdout_contains: "hello"
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_SuccessfulExecution")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "All tests passed") {
		t.Errorf("Expected 'All tests passed' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "PASSED") {
		t.Errorf("Expected 'PASSED' in output, got: %s", outputStr)
	}
}

// TestMain_FailingTests tests the main function with failing tests (exit code 1).
func TestMain_FailingTests(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_failure.yaml")
		configContent := `tests:
  - name: test-failure
    command: "echo wrong"
    expect:
      exit_code: 0
      stdout_contains: "expected"
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_FailingTests")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 1, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Completed with 1 failing test(s)") {
		t.Errorf("Expected 'Completed with 1 failing test(s)' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "FAILED") {
		t.Errorf("Expected 'FAILED' in output, got: %s", outputStr)
	}
}

// TestMain_ConfigLoadFailure tests the main function with config loading failure (exit code 2).
func TestMain_ConfigLoadFailure(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		os.Args = []string{"cmd", "-config", "/nonexistent/config.yaml", "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_ConfigLoadFailure")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 2, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error: failed to load tests:") {
		t.Errorf("Expected 'Error: failed to load tests:' in output, got: %s", outputStr)
	}
}

// TestMain_InvalidConfigFormat tests the main function with invalid YAML format (exit code 2).
func TestMain_InvalidConfigFormat(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_invalid.yaml")
		configContent := `tests:
  - name: test
    invalid yaml content [[[
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_InvalidConfigFormat")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 2, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error: failed to load tests:") {
		t.Errorf("Expected 'Error: failed to load tests:' in output, got: %s", outputStr)
	}
}

// TestMain_JSONReportSuccess tests the main function with JSON report generation.
func TestMain_JSONReportSuccess(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_json.yaml")
		configContent := `tests:
  - name: test-json
    command: "echo test"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		reportFile := filepath.Join(tmpDir, "report.json")
		defer os.Remove(reportFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-json-report", reportFile}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_JSONReportSuccess")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Report written to") {
		t.Errorf("Expected 'Report written to' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "report.json") {
		t.Errorf("Expected 'report.json' in output, got: %s", outputStr)
	}
}

// TestMain_JSONReportWithFailures tests JSON report generation with test failures.
func TestMain_JSONReportWithFailures(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_json_fail.yaml")
		configContent := `tests:
  - name: test-json-fail
    command: "echo wrong"
    expect:
      exit_code: 1
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		reportFile := filepath.Join(tmpDir, "report_fail.json")
		defer os.Remove(reportFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-json-report", reportFile}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_JSONReportWithFailures")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 1, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Report written to") {
		t.Errorf("Expected 'Report written to' in output, got: %s", outputStr)
	}
}

// TestMain_JSONReportWriteFailure tests JSON report write failure (exit code 2).
func TestMain_JSONReportWriteFailure(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_json_write_fail.yaml")
		configContent := `tests:
  - name: test-json-write-fail
    command: "echo test"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		// Use invalid path to trigger write failure
		reportFile := "/nonexistent/directory/report.json"

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-json-report", reportFile}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_JSONReportWriteFailure")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 2, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error: failed to write report:") {
		t.Errorf("Expected 'Error: failed to write report:' in output, got: %s", outputStr)
	}
}

// TestMain_MultipleTests tests the main function with multiple tests.
func TestMain_MultipleTests(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_multiple.yaml")
		configContent := `tests:
  - name: test-1
    command: "echo test1"
    expect:
      exit_code: 0
  - name: test-2
    command: "echo test2"
    expect:
      exit_code: 0
  - name: test-3
    command: "echo test3"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MultipleTests")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "test-1") {
		t.Errorf("Expected 'test-1' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "test-2") {
		t.Errorf("Expected 'test-2' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "test-3") {
		t.Errorf("Expected 'test-3' in output, got: %s", outputStr)
	}
}

// TestMain_FailFast tests the main function with fail-fast mode.
func TestMain_FailFast(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_failfast.yaml")
		configContent := `tests:
  - name: test-1
    command: "echo wrong"
    expect:
      stdout_contains: "expected"
  - name: test-2
    command: "echo test2"
    expect:
      exit_code: 0
  - name: test-3
    command: "echo test3"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-fail-fast"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_FailFast")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 1, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Stopping due to fail-fast") {
		t.Errorf("Expected 'Stopping due to fail-fast' in output, got: %s", outputStr)
	}
	// test-2 and test-3 should not run
	if strings.Contains(outputStr, "test-2") || strings.Contains(outputStr, "test-3") {
		t.Errorf("Expected test-2 and test-3 not to run with fail-fast, got: %s", outputStr)
	}
}

// TestMain_DryRun tests the main function with dry-run mode.
func TestMain_DryRun(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_dryrun.yaml")
		configContent := `tests:
  - name: test-dryrun
    command: "echo test"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-dry-run"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_DryRun")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "DRY-RUN") {
		t.Errorf("Expected 'DRY-RUN' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "All tests passed") {
		t.Errorf("Expected 'All tests passed' in output, got: %s", outputStr)
	}
}

// TestMain_DebugMode tests the main function with debug mode enabled.
func TestMain_DebugMode(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_debug.yaml")
		configContent := `tests:
  - name: test-debug
    command: "echo debug"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-debug"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_DebugMode")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "[debug]") {
		t.Errorf("Expected '[debug]' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "full command:") {
		t.Errorf("Expected 'full command:' in debug output, got: %s", outputStr)
	}
}

// TestMain_SkippedTests tests the main function with skipped tests.
func TestMain_SkippedTests(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_skip.yaml")
		configContent := `tests:
  - name: test-skipped
    command: "echo test"
    skip: true
    expect:
      exit_code: 0
  - name: test-run
    command: "echo test"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_SkippedTests")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "SKIPPED") {
		t.Errorf("Expected 'SKIPPED' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "All tests passed") {
		t.Errorf("Expected 'All tests passed' in output, got: %s", outputStr)
	}
}

// TestMain_JSONReportVerification verifies the actual JSON report content.
func TestMain_JSONReportVerification(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_json_verify.yaml")
		configContent := `tests:
  - name: test-verify
    command: "echo verify"
    expect:
      exit_code: 0
      stdout_contains: "verify"
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		reportFile := filepath.Join(tmpDir, "verify_report.json")

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-json-report", reportFile}
		main()
		return
	}

	tmpDir := os.TempDir()
	reportFile := filepath.Join(tmpDir, "verify_report.json")
	defer os.Remove(reportFile)

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_JSONReportVerification")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	// Verify JSON file was created and contains valid JSON
	if _, err := os.Stat(reportFile); os.IsNotExist(err) {
		t.Fatal("JSON report file was not created")
	}

	data, err := os.ReadFile(reportFile)
	if err != nil {
		t.Fatalf("Failed to read JSON report: %v", err)
	}

	var results []runner.Result
	if err := json.Unmarshal(data, &results); err != nil {
		t.Fatalf("Failed to unmarshal JSON report: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result in JSON report, got %d", len(results))
	}

	if results[0].Status != "PASSED" {
		t.Errorf("Expected status 'PASSED', got '%s'", results[0].Status)
	}

	if results[0].Name != "test-verify" {
		t.Errorf("Expected name 'test-verify', got '%s'", results[0].Name)
	}
}

// TestMain_MixedResultsCount tests counting of passed and failed tests.
func TestMain_MixedResultsCount(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_mixed.yaml")
		configContent := `tests:
  - name: test-pass-1
    command: "echo pass"
    expect:
      exit_code: 0
  - name: test-fail-1
    command: "echo fail"
    expect:
      stdout_contains: "expected"
  - name: test-pass-2
    command: "echo pass2"
    expect:
      exit_code: 0
  - name: test-fail-2
    command: "echo fail2"
    expect:
      stdout_contains: "expected"
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MixedResultsCount")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatal("Expected exit code 1, but got no error")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Completed with 2 failing test(s)") {
		t.Errorf("Expected 'Completed with 2 failing test(s)' in output, got: %s", outputStr)
	}
}

// TestMain_CustomEngine tests the main function with a custom engine.
func TestMain_CustomEngine(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_engine.yaml")
		configContent := `tests:
  - name: test-podman
    command: "echo podman"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		// Use docker even if named podman since we're just testing the flag is passed
		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-engine", "docker"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_CustomEngine")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "All tests passed") {
		t.Errorf("Expected 'All tests passed' in output, got: %s", outputStr)
	}
}

// TestMain_CustomTimeout tests the main function with custom default timeout.
func TestMain_CustomTimeout(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		tmpDir := os.TempDir()
		configFile := filepath.Join(tmpDir, "test_timeout.yaml")
		configContent := `tests:
  - name: test-timeout
    command: "echo timeout"
    expect:
      exit_code: 0
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer os.Remove(configFile)

		os.Args = []string{"cmd", "-config", configFile, "-image", "alpine:latest", "-default-timeout", "60"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_CustomTimeout")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Expected exit code 0, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "All tests passed") {
		t.Errorf("Expected 'All tests passed' in output, got: %s", outputStr)
	}
}
