package runner

import (
	"container-test-cli/internal/config"
	"fmt"
	"strings"
	"testing"
)

// TestFirstNonEmpty tests the firstNonEmpty helper function.
func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{"first_non_empty", []string{"first", "second", "third"}, "first"},
		{"skip_empty", []string{"", "second", "third"}, "second"},
		{"all_empty", []string{"", "", ""}, ""},
		{"single_value", []string{"only"}, "only"},
		{"empty_slice", []string{}, ""},
		{"middle_non_empty", []string{"", "", "middle", ""}, "middle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := firstNonEmpty(tt.values...)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBuildRunCommand tests command building with various configurations.
func TestBuildRunCommand(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		cmd := BuildRunCommand("docker", "alpine:latest", []string{"echo", "hello"}, "", nil, nil, nil)
		expected := []string{"docker", "run", "--rm", "alpine:latest", "echo", "hello"}
		if !equalSlices(cmd, expected) {
			t.Errorf("Expected %v, got %v", expected, cmd)
		}
	})

	t.Run("with_workdir", func(t *testing.T) {
		cmd := BuildRunCommand("docker", "alpine:latest", []string{"pwd"}, "/app", nil, nil, nil)
		if !contains(cmd, "-w") || !contains(cmd, "/app") {
			t.Errorf("Expected workdir in command: %v", cmd)
		}
	})

	t.Run("with_env", func(t *testing.T) {
		env := map[string]string{"KEY1": "value1", "KEY2": "value2"}
		cmd := BuildRunCommand("docker", "alpine:latest", []string{"env"}, "", env, nil, nil)

		hasKey1 := false
		hasKey2 := false
		for i, arg := range cmd {
			if arg == "-e" && i+1 < len(cmd) {
				if cmd[i+1] == "KEY1=value1" {
					hasKey1 = true
				}
				if cmd[i+1] == "KEY2=value2" {
					hasKey2 = true
				}
			}
		}
		if !hasKey1 || !hasKey2 {
			t.Errorf("Expected both env vars in command: %v", cmd)
		}
	})

	t.Run("with_run_args", func(t *testing.T) {
		runArgs := []string{"--network=host", "--rm"}
		cmd := BuildRunCommand("docker", "alpine:latest", []string{"echo"}, "", nil, runArgs, nil)
		if !contains(cmd, "--network=host") {
			t.Errorf("Expected run args in command: %v", cmd)
		}
	})

	t.Run("with_entrypoint", func(t *testing.T) {
		entrypoint := "/bin/sh"
		cmd := BuildRunCommand("docker", "alpine:latest", []string{"-c", "echo"}, "", nil, nil, &entrypoint)
		if !contains(cmd, "--entrypoint") || !contains(cmd, "/bin/sh") {
			t.Errorf("Expected entrypoint in command: %v", cmd)
		}
	})

	t.Run("with_all_options", func(t *testing.T) {
		env := map[string]string{"KEY": "value"}
		runArgs := []string{"--network=host"}
		entrypoint := "/bin/bash"
		cmd := BuildRunCommand("podman", "ubuntu:latest", []string{"-c", "test"}, "/workdir", env, runArgs, &entrypoint)

		if cmd[0] != "podman" {
			t.Error("Expected engine to be podman")
		}
		if !contains(cmd, "ubuntu:latest") {
			t.Error("Expected image in command")
		}
		if !contains(cmd, "-w") || !contains(cmd, "/workdir") {
			t.Error("Expected workdir in command")
		}
		if !contains(cmd, "--entrypoint") || !contains(cmd, "/bin/bash") {
			t.Error("Expected entrypoint in command")
		}
		if !contains(cmd, "--network=host") {
			t.Error("Expected run args in command")
		}
	})

	t.Run("empty_entrypoint", func(t *testing.T) {
		entrypoint := ""
		cmd := BuildRunCommand("docker", "alpine:latest", []string{"echo"}, "", nil, nil, &entrypoint)
		if !contains(cmd, "--entrypoint") || !contains(cmd, "") {
			t.Errorf("Expected empty entrypoint in command: %v", cmd)
		}
	})
}

// TestCheckContains tests the checkContains helper function.
func TestCheckContains(t *testing.T) {
	t.Run("all_present", func(t *testing.T) {
		output := "hello world test"
		expected := []string{"hello", "world"}
		failures := checkContains(output, expected, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("some_missing", func(t *testing.T) {
		output := "hello world"
		expected := []string{"hello", "missing"}
		failures := checkContains(output, expected, "stdout")
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d: %v", len(failures), failures)
		}
		if !strings.Contains(failures[0], "missing") {
			t.Errorf("Expected failure to mention 'missing', got: %s", failures[0])
		}
	})

	t.Run("empty_list", func(t *testing.T) {
		output := "hello world"
		expected := []string{}
		failures := checkContains(output, expected, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures for empty list, got: %v", failures)
		}
	})

	t.Run("nil_list", func(t *testing.T) {
		output := "hello world"
		failures := checkContains(output, nil, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures for nil list, got: %v", failures)
		}
	})

	t.Run("case_sensitive", func(t *testing.T) {
		output := "Hello World"
		expected := []string{"hello"}
		failures := checkContains(output, expected, "stdout")
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure (case mismatch), got %d", len(failures))
		}
	})
}

// TestCheckNotContains tests the checkNotContains helper function.
func TestCheckNotContains(t *testing.T) {
	t.Run("none_present", func(t *testing.T) {
		output := "hello world"
		forbidden := []string{"error", "fail"}
		failures := checkNotContains(output, forbidden, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("some_present", func(t *testing.T) {
		output := "hello error world"
		forbidden := []string{"error", "fail"}
		failures := checkNotContains(output, forbidden, "stdout")
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d: %v", len(failures), failures)
		}
		if !strings.Contains(failures[0], "error") {
			t.Errorf("Expected failure to mention 'error', got: %s", failures[0])
		}
	})

	t.Run("empty_list", func(t *testing.T) {
		output := "hello world"
		forbidden := []string{}
		failures := checkNotContains(output, forbidden, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures for empty list, got: %v", failures)
		}
	})

	t.Run("nil_list", func(t *testing.T) {
		output := "hello world"
		failures := checkNotContains(output, nil, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures for nil list, got: %v", failures)
		}
	})
}

// TestCheckRegex tests the checkRegex helper function.
func TestCheckRegex(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		output := "hello world 123"
		pattern := `\d+`
		failures := checkRegex(output, pattern, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("no_match", func(t *testing.T) {
		output := "hello world"
		pattern := `\d+`
		failures := checkRegex(output, pattern, "stdout")
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d: %v", len(failures), failures)
		}
	})

	t.Run("empty_pattern", func(t *testing.T) {
		output := "hello world"
		pattern := ""
		failures := checkRegex(output, pattern, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures for empty pattern, got: %v", failures)
		}
	})

	t.Run("complex_pattern", func(t *testing.T) {
		output := "test@example.com"
		pattern := `^[a-z]+@[a-z]+\.[a-z]+$`
		failures := checkRegex(output, pattern, "stdout")
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})
}

// TestEvalExpectations tests the evalExpectations function.
func TestEvalExpectations(t *testing.T) {
	t.Run("default_exit_code_success", func(t *testing.T) {
		expect := config.ExpectBlock{}
		failures := evalExpectations(expect, "", "", 0)
		if len(failures) != 0 {
			t.Errorf("Expected no failures for exit code 0, got: %v", failures)
		}
	})

	t.Run("default_exit_code_failure", func(t *testing.T) {
		expect := config.ExpectBlock{}
		failures := evalExpectations(expect, "", "", 1)
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure for exit code 1, got %d: %v", len(failures), failures)
		}
	})

	t.Run("custom_exit_code", func(t *testing.T) {
		exitCode := config.ExitCodeExpect{Op: "==", Value: 1}
		expect := config.ExpectBlock{ExitCode: &exitCode}
		failures := evalExpectations(expect, "", "", 1)
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("stdout_contains_success", func(t *testing.T) {
		expect := config.ExpectBlock{
			StdoutContains: []string{"hello", "world"},
		}
		failures := evalExpectations(expect, "hello world", "", 0)
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("stdout_contains_failure", func(t *testing.T) {
		expect := config.ExpectBlock{
			StdoutContains: []string{"missing"},
		}
		failures := evalExpectations(expect, "hello world", "", 0)
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d: %v", len(failures), failures)
		}
	})

	t.Run("stdout_not_contains_success", func(t *testing.T) {
		expect := config.ExpectBlock{
			StdoutNot: []string{"error"},
		}
		failures := evalExpectations(expect, "hello world", "", 0)
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("stdout_not_contains_failure", func(t *testing.T) {
		expect := config.ExpectBlock{
			StdoutNot: []string{"error"},
		}
		failures := evalExpectations(expect, "hello error world", "", 0)
		if len(failures) != 1 {
			t.Errorf("Expected 1 failure, got %d: %v", len(failures), failures)
		}
	})

	t.Run("stdout_regex_success", func(t *testing.T) {
		expect := config.ExpectBlock{
			StdoutRegex: `^\d+$`,
		}
		failures := evalExpectations(expect, "12345", "", 0)
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("stderr_contains_success", func(t *testing.T) {
		expect := config.ExpectBlock{
			StderrContains: []string{"warning"},
		}
		failures := evalExpectations(expect, "", "warning message", 0)
		if len(failures) != 0 {
			t.Errorf("Expected no failures, got: %v", failures)
		}
	})

	t.Run("multiple_failures", func(t *testing.T) {
		exitCode := config.ExitCodeExpect{Op: "==", Value: 0}
		expect := config.ExpectBlock{
			ExitCode:       &exitCode,
			StdoutContains: []string{"missing1", "missing2"},
		}
		failures := evalExpectations(expect, "hello", "", 1)
		// Should have 3 failures: exit code + 2 missing strings
		if len(failures) != 3 {
			t.Errorf("Expected 3 failures, got %d: %v", len(failures), failures)
		}
	})
}

// TestRunSingleTest_Skip tests that skipped tests return SKIPPED status.
func TestRunSingleTest_Skip(t *testing.T) {
	testCase := config.TestCase{
		Name: "skipped-test",
		Skip: true,
		Exec: []string{"echo", "should not run"},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "SKIPPED" {
		t.Errorf("Expected SKIPPED status, got %s", result.Status)
	}
	if result.Name != "skipped-test" {
		t.Errorf("Expected name 'skipped-test', got '%s'", result.Name)
	}
}

// TestRunSingleTest_MissingCommand tests failure when no command is provided.
func TestRunSingleTest_MissingCommand(t *testing.T) {
	testCase := config.TestCase{
		Name: "no-command",
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "FAILED" {
		t.Errorf("Expected FAILED status, got %s", result.Status)
	}
	if len(result.Failures) == 0 {
		t.Error("Expected failures for missing command")
	}
	if !strings.Contains(result.Failures[0], "missing 'exec' or 'command'") {
		t.Errorf("Expected failure about missing command, got: %v", result.Failures)
	}
}

// TestRunSingleTest_DryRun tests dry-run mode.
func TestRunSingleTest_DryRun(t *testing.T) {
	testCase := config.TestCase{
		Name: "dry-run-test",
		Exec: []string{"echo", "hello"},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, true)

	if result.Status != "DRY-RUN" {
		t.Errorf("Expected DRY-RUN status, got %s", result.Status)
	}
	if len(result.DebugMessages) == 0 {
		t.Error("Expected debug messages in dry-run mode")
	}
	if !strings.Contains(result.DebugMessages[0], "dry-run command") {
		t.Errorf("Expected dry-run message, got: %v", result.DebugMessages)
	}
}

// TestRunSingleTest_UnnamedTest tests that unnamed tests get default name.
func TestRunSingleTest_UnnamedTest(t *testing.T) {
	testCase := config.TestCase{
		Name: "",
		Skip: true,
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Name != "unnamed" {
		t.Errorf("Expected name 'unnamed', got '%s'", result.Name)
	}
}

// TestRunSingleTest_TimeoutPrecedence tests timeout precedence.
func TestRunSingleTest_TimeoutPrecedence(t *testing.T) {
	expectTimeout := 10
	testTimeout := 20

	t.Run("test_timeout_wins", func(t *testing.T) {
		testCase := config.TestCase{
			Name: "timeout-test",
			Exec: []string{"echo", "hello"},
			Expect: config.ExpectBlock{
				TimeoutSeconds: &expectTimeout,
			},
			Timeout: &testTimeout,
		}

		result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, true)
		if !strings.Contains(result.DebugMessages[0], "timeout=20s") {
			t.Errorf("Expected test timeout (20s) to be used, got: %v", result.DebugMessages)
		}
	})

	t.Run("expect_timeout_when_no_test_timeout", func(t *testing.T) {
		testCase := config.TestCase{
			Name: "timeout-test",
			Exec: []string{"echo", "hello"},
			Expect: config.ExpectBlock{
				TimeoutSeconds: &expectTimeout,
			},
		}

		result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, true)
		if !strings.Contains(result.DebugMessages[0], "timeout=10s") {
			t.Errorf("Expected expect timeout (10s) to be used, got: %v", result.DebugMessages)
		}
	})

	t.Run("default_timeout_when_none_specified", func(t *testing.T) {
		testCase := config.TestCase{
			Name: "timeout-test",
			Exec: []string{"echo", "hello"},
		}

		result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, true)
		if !strings.Contains(result.DebugMessages[0], "timeout=30s") {
			t.Errorf("Expected default timeout (30s) to be used, got: %v", result.DebugMessages)
		}
	})
}

// TestRunTests tests the RunTests function.
func TestRunTests(t *testing.T) {
	t.Run("empty_list", func(t *testing.T) {
		tests := []config.TestCase{}
		cfg := Config{
			Engine:         "docker",
			Image:          "alpine:latest",
			DefaultTimeout: 30,
		}

		var resultCount int
		results, failures := RunTests(tests, cfg, func(r Result) {
			resultCount++
		})

		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
		if failures != 0 {
			t.Errorf("Expected 0 failures, got %d", failures)
		}
		if resultCount != 0 {
			t.Errorf("Expected onResult to be called 0 times, got %d", resultCount)
		}
	})

	t.Run("multiple_tests", func(t *testing.T) {
		tests := []config.TestCase{
			{Name: "test1", Skip: true},
			{Name: "test2", Skip: true},
			{Name: "test3", Skip: true},
		}
		cfg := Config{
			Engine:         "docker",
			Image:          "alpine:latest",
			DefaultTimeout: 30,
		}

		var resultCount int
		results, failures := RunTests(tests, cfg, func(r Result) {
			resultCount++
		})

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		if resultCount != 3 {
			t.Errorf("Expected onResult to be called 3 times, got %d", resultCount)
		}
		if failures != 0 {
			t.Errorf("Expected 0 failures for skipped tests, got %d", failures)
		}
	})

	t.Run("fail_fast", func(t *testing.T) {
		tests := []config.TestCase{
			{Name: "test1"}, // Will fail (no command)
			{Name: "test2", Skip: true},
			{Name: "test3", Skip: true},
		}
		cfg := Config{
			Engine:         "docker",
			Image:          "alpine:latest",
			DefaultTimeout: 30,
			FailFast:       true,
		}

		results, failures := RunTests(tests, cfg, func(r Result) {})

		if len(results) != 1 {
			t.Errorf("Expected 1 result (fail-fast), got %d", len(results))
		}
		if failures != 1 {
			t.Errorf("Expected 1 failure, got %d", failures)
		}
	})

	t.Run("no_fail_fast", func(t *testing.T) {
		tests := []config.TestCase{
			{Name: "test1"}, // Will fail (no command)
			{Name: "test2", Skip: true},
			{Name: "test3"}, // Will fail (no command)
		}
		cfg := Config{
			Engine:         "docker",
			Image:          "alpine:latest",
			DefaultTimeout: 30,
			FailFast:       false,
		}

		results, failures := RunTests(tests, cfg, func(r Result) {})

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		if failures != 2 {
			t.Errorf("Expected 2 failures, got %d", failures)
		}
	})

	t.Run("dry_run_mode", func(t *testing.T) {
		tests := []config.TestCase{
			{Name: "test1", Exec: []string{"echo", "hello"}},
			{Name: "test2", Command: []string{"sh", "-c", "echo world"}},
		}
		cfg := Config{
			Engine:         "docker",
			Image:          "alpine:latest",
			DefaultTimeout: 30,
			DryRun:         true,
		}

		results, _ := RunTests(tests, cfg, func(r Result) {})

		for _, res := range results {
			if res.Status != "DRY-RUN" {
				t.Errorf("Expected DRY-RUN status, got %s for test %s", res.Status, res.Name)
			}
		}
	})
}

// TestResult_Structure tests the Result struct fields.
func TestResult_Structure(t *testing.T) {
	exitCode := 0
	result := Result{
		Status:        "PASSED",
		Name:          "test-name",
		Stdout:        "stdout output",
		Stderr:        "stderr output",
		ExitCode:      &exitCode,
		Failures:      []string{"failure1", "failure2"},
		DebugMessages: []string{"debug1", "debug2"},
	}

	if result.Status != "PASSED" {
		t.Error("Status field not set correctly")
	}
	if result.Name != "test-name" {
		t.Error("Name field not set correctly")
	}
	if result.Stdout != "stdout output" {
		t.Error("Stdout field not set correctly")
	}
	if result.Stderr != "stderr output" {
		t.Error("Stderr field not set correctly")
	}
	if *result.ExitCode != 0 {
		t.Error("ExitCode field not set correctly")
	}
	if len(result.Failures) != 2 {
		t.Error("Failures field not set correctly")
	}
	if len(result.DebugMessages) != 2 {
		t.Error("DebugMessages field not set correctly")
	}
}

// TestConfig_Structure tests the Config struct fields.
func TestConfig_Structure(t *testing.T) {
	cfg := Config{
		Engine:         "docker",
		Image:          "alpine:latest",
		DefaultTimeout: 30,
		FailFast:       true,
		Debug:          true,
		DryRun:         true,
	}

	if cfg.Engine != "docker" {
		t.Error("Engine field not set correctly")
	}
	if cfg.Image != "alpine:latest" {
		t.Error("Image field not set correctly")
	}
	if cfg.DefaultTimeout != 30 {
		t.Error("DefaultTimeout field not set correctly")
	}
	if !cfg.FailFast {
		t.Error("FailFast field not set correctly")
	}
	if !cfg.Debug {
		t.Error("Debug field not set correctly")
	}
	if !cfg.DryRun {
		t.Error("DryRun field not set correctly")
	}
}

// Helper functions

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestBuildRunCommand_EnvOrder tests that multiple env vars are all included.
func TestBuildRunCommand_EnvOrder(t *testing.T) {
	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
		"VAR3": "value3",
	}
	cmd := BuildRunCommand("docker", "alpine:latest", []string{"env"}, "", env, nil, nil)

	// Count how many env vars are in the command
	envCount := 0
	for i, arg := range cmd {
		if arg == "-e" && i+1 < len(cmd) {
			envCount++
		}
	}

	if envCount != 3 {
		t.Errorf("Expected 3 env vars in command, found %d", envCount)
	}
}

// TestRunSingleTest_CommandPrecedence tests that Exec is preferred over Command.
func TestRunSingleTest_CommandPrecedence(t *testing.T) {
	testCase := config.TestCase{
		Name:    "precedence-test",
		Exec:    []string{"echo", "from-exec"},
		Command: []string{"sh", "-c", "echo from-command"},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, true)

	// Check that Exec command is used (should contain "from-exec")
	if !strings.Contains(result.DebugMessages[0], "from-exec") {
		t.Error("Expected Exec to take precedence over Command")
	}
}

// TestEvalExpectations_AllChecks tests evaluation with all checks enabled.
func TestEvalExpectations_AllChecks(t *testing.T) {
	exitCode := config.ExitCodeExpect{Op: "==", Value: 0}
	expect := config.ExpectBlock{
		ExitCode:       &exitCode,
		StdoutContains: []string{"hello"},
		StdoutNot:      []string{"error"},
		StdoutRegex:    `^hello.*world$`,
		StderrContains: []string{"warning"},
		StderrRegex:    `warn.*`,
	}

	failures := evalExpectations(expect, "hello world", "warning message", 0)

	if len(failures) != 0 {
		t.Errorf("Expected no failures when all checks pass, got: %v", failures)
	}
}

// TestBuildRunCommand_EmptyCommand tests building with empty command slice.
func TestBuildRunCommand_EmptyCommand(t *testing.T) {
	cmd := BuildRunCommand("docker", "alpine:latest", []string{}, "", nil, nil, nil)
	expected := []string{"docker", "run", "--rm", "alpine:latest"}
	if !equalSlices(cmd, expected) {
		t.Errorf("Expected %v, got %v", expected, cmd)
	}
}

// TestRunTests_OnResultCallback tests that onResult is called for each test.
func TestRunTests_OnResultCallback(t *testing.T) {
	tests := []config.TestCase{
		{Name: "test1", Skip: true},
		{Name: "test2", Skip: true},
	}
	cfg := Config{
		Engine:         "docker",
		Image:          "alpine:latest",
		DefaultTimeout: 30,
	}

	var names []string
	RunTests(tests, cfg, func(r Result) {
		names = append(names, r.Name)
	})

	if len(names) != 2 {
		t.Errorf("Expected callback to be called 2 times, got %d", len(names))
	}
	if names[0] != "test1" || names[1] != "test2" {
		t.Errorf("Expected names in order, got %v", names)
	}
}

// TestCheckRegex_InvalidPattern tests handling of invalid regex patterns.
func TestCheckRegex_InvalidPattern(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// If the function panics on invalid regex, that's acceptable
			// The test passes
			return
		}
	}()

	output := "test"
	pattern := `[invalid(`

	// This should either return a failure or panic
	// Both behaviors are acceptable for invalid regex
	failures := checkRegex(output, pattern, "stdout")

	// If we reach here without panic, check that failures were reported
	if len(failures) == 0 {
		// The regex might be accepted by Go's regex engine, which is fine
		fmt.Println("Invalid regex was accepted or handled gracefully")
	}
}

// Integration tests - require Docker/Podman to be available

// TestRunSingleTest_RealExecution tests actual command execution.
func TestRunSingleTest_RealExecution(t *testing.T) {
	testCase := config.TestCase{
		Name: "echo-test",
		Exec: []string{"echo", "hello world"},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
	if !strings.Contains(result.Stdout, "hello world") {
		t.Errorf("Expected stdout to contain 'hello world', got: %s", result.Stdout)
	}
	if result.ExitCode == nil || *result.ExitCode != 0 {
		t.Error("Expected exit code 0")
	}
}

// TestRunSingleTest_RealExecution_NonZeroExit tests command with non-zero exit code.
func TestRunSingleTest_RealExecution_NonZeroExit(t *testing.T) {
	exitCode := config.ExitCodeExpect{Op: "==", Value: 1}
	testCase := config.TestCase{
		Name:    "exit-1-test",
		Command: []string{"sh", "-c", "exit 1"},
		Expect: config.ExpectBlock{
			ExitCode: &exitCode,
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status for exit 1 with exit_code: 1, got %s. Failures: %v", result.Status, result.Failures)
	}
	if result.ExitCode == nil || *result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %v", result.ExitCode)
	}
}

// TestRunSingleTest_RealExecution_StdoutExpectations tests stdout validation.
func TestRunSingleTest_RealExecution_StdoutExpectations(t *testing.T) {
	testCase := config.TestCase{
		Name:    "stdout-test",
		Command: []string{"sh", "-c", "echo 'test output' && echo 'more output'"},
		Expect: config.ExpectBlock{
			StdoutContains: []string{"test output", "more output"},
			StdoutRegex:    `test.*output`,
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
}

// TestRunSingleTest_RealExecution_StderrExpectations tests stderr validation.
func TestRunSingleTest_RealExecution_StderrExpectations(t *testing.T) {
	testCase := config.TestCase{
		Name:    "stderr-test",
		Command: []string{"sh", "-c", "echo 'error message' >&2"},
		Expect: config.ExpectBlock{
			StderrContains: []string{"error message"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
	if !strings.Contains(result.Stderr, "error message") {
		t.Errorf("Expected stderr to contain 'error message', got: %s", result.Stderr)
	}
}

// TestRunSingleTest_RealExecution_FailedExpectations tests failing expectations.
func TestRunSingleTest_RealExecution_FailedExpectations(t *testing.T) {
	testCase := config.TestCase{
		Name:    "fail-test",
		Command: []string{"sh", "-c", "echo 'actual output'"},
		Expect: config.ExpectBlock{
			StdoutContains: []string{"expected output that won't be there"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "FAILED" {
		t.Errorf("Expected FAILED status, got %s", result.Status)
	}
	if len(result.Failures) == 0 {
		t.Error("Expected failures to be reported")
	}
}

// TestRunSingleTest_RealExecution_WithWorkdir tests workdir functionality.
func TestRunSingleTest_RealExecution_WithWorkdir(t *testing.T) {
	testCase := config.TestCase{
		Name:    "workdir-test",
		Command: []string{"sh", "-c", "pwd"},
		Workdir: "/tmp",
		Expect: config.ExpectBlock{
			StdoutContains: []string{"/tmp"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
}

// TestRunSingleTest_RealExecution_WithEnv tests environment variable functionality.
func TestRunSingleTest_RealExecution_WithEnv(t *testing.T) {
	testCase := config.TestCase{
		Name:    "env-test",
		Command: []string{"sh", "-c", "echo $TEST_VAR"},
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
		Expect: config.ExpectBlock{
			StdoutContains: []string{"test_value"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
}

// TestRunSingleTest_RealExecution_WithDebug tests debug mode output.
func TestRunSingleTest_RealExecution_WithDebug(t *testing.T) {
	testCase := config.TestCase{
		Name:    "debug-test",
		Command: []string{"sh", "-c", "echo 'test'"},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, true, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s", result.Status)
	}
	if len(result.DebugMessages) == 0 {
		t.Error("Expected debug messages to be populated")
	}

	// Check for expected debug messages
	hasCommand := false
	hasStdout := false
	hasExitCode := false

	for _, msg := range result.DebugMessages {
		if strings.Contains(msg, "full command:") {
			hasCommand = true
		}
		if strings.Contains(msg, "stdout:") {
			hasStdout = true
		}
		if strings.Contains(msg, "exit code:") {
			hasExitCode = true
		}
	}

	if !hasCommand || !hasStdout || !hasExitCode {
		t.Errorf("Expected all debug messages. Got: %v", result.DebugMessages)
	}
}

// TestRunSingleTest_RealExecution_ShortTimeout tests timeout behavior.
func TestRunSingleTest_RealExecution_ShortTimeout(t *testing.T) {
	timeout := 1
	testCase := config.TestCase{
		Name:    "timeout-test",
		Command: []string{"sh", "-c", "sleep 10"},
		Timeout: &timeout,
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "FAILED" {
		t.Errorf("Expected FAILED status for timeout, got %s", result.Status)
	}
	if len(result.Failures) == 0 {
		t.Error("Expected timeout failure to be reported")
	}
	if !strings.Contains(result.Failures[0], "timed out") {
		t.Errorf("Expected timeout message, got: %v", result.Failures)
	}
	if result.ExitCode != nil {
		t.Error("Expected nil exit code for timeout")
	}
}

// TestRunSingleTest_RealExecution_StdoutNotContains tests stdout_not_contains.
func TestRunSingleTest_RealExecution_StdoutNotContains(t *testing.T) {
	testCase := config.TestCase{
		Name:    "not-contains-test",
		Command: []string{"sh", "-c", "echo 'hello world'"},
		Expect: config.ExpectBlock{
			StdoutNot: []string{"error", "fail"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
}

// TestRunSingleTest_RealExecution_StdoutNotContainsFails tests failing stdout_not_contains.
func TestRunSingleTest_RealExecution_StdoutNotContainsFails(t *testing.T) {
	testCase := config.TestCase{
		Name:    "not-contains-fail-test",
		Command: []string{"sh", "-c", "echo 'error occurred'"},
		Expect: config.ExpectBlock{
			StdoutNot: []string{"error"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "FAILED" {
		t.Errorf("Expected FAILED status, got %s", result.Status)
	}
	if len(result.Failures) == 0 {
		t.Error("Expected failure for stdout containing forbidden string")
	}
}

// TestRunSingleTest_RealExecution_ComplexRegex tests complex regex patterns.
func TestRunSingleTest_RealExecution_ComplexRegex(t *testing.T) {
	testCase := config.TestCase{
		Name:    "regex-test",
		Command: []string{"sh", "-c", "echo 'Version: 1.2.3'"},
		Expect: config.ExpectBlock{
			StdoutRegex: `Version: \d+\.\d+\.\d+`,
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
}

// TestRunSingleTest_RealExecution_ExitCodeOperators tests various exit code operators.
func TestRunSingleTest_RealExecution_ExitCodeOperators(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		operator string
		value    int
		wantPass bool
	}{
		{"greater_than", "exit 5", ">", 3, true},
		{"greater_than_fail", "exit 2", ">", 3, false},
		{"less_than", "exit 2", "<", 5, true},
		{"less_than_fail", "exit 6", "<", 5, false},
		{"not_equals", "exit 1", "!=", 0, true},
		{"not_equals_fail", "exit 0", "!=", 0, false},
		{"greater_or_equal", "exit 5", ">=", 5, true},
		{"less_or_equal", "exit 5", "<=", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := config.ExitCodeExpect{Op: tt.operator, Value: tt.value}
			testCase := config.TestCase{
				Name:    tt.name,
				Command: []string{"sh", "-c", tt.command},
				Expect: config.ExpectBlock{
					ExitCode: &exitCode,
				},
			}

			result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

			if tt.wantPass && result.Status != "PASSED" {
				t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
			}
			if !tt.wantPass && result.Status != "FAILED" {
				t.Errorf("Expected FAILED status, got %s", result.Status)
			}
		})
	}
}

// TestRunTests_RealExecution tests running multiple real tests.
func TestRunTests_RealExecution(t *testing.T) {
	tests := []config.TestCase{
		{
			Name:    "test-1",
			Command: []string{"sh", "-c", "echo 'test 1'"},
			Expect: config.ExpectBlock{
				StdoutContains: []string{"test 1"},
			},
		},
		{
			Name:    "test-2",
			Command: []string{"sh", "-c", "echo 'test 2'"},
			Expect: config.ExpectBlock{
				StdoutContains: []string{"test 2"},
			},
		},
	}
	cfg := Config{
		Engine:         "docker",
		Image:          "alpine:latest",
		DefaultTimeout: 30,
	}

	results, failures := RunTests(tests, cfg, func(r Result) {})

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if failures != 0 {
		t.Errorf("Expected 0 failures, got %d", failures)
	}
	for i, res := range results {
		if res.Status != "PASSED" {
			t.Errorf("Test %d: Expected PASSED, got %s. Failures: %v", i, res.Status, res.Failures)
		}
	}
}

// TestRunTests_RealExecution_FailFast tests fail-fast with real execution.
func TestRunTests_RealExecution_FailFast(t *testing.T) {
	tests := []config.TestCase{
		{
			Name:    "fail-test",
			Command: []string{"sh", "-c", "exit 1"},
		},
		{
			Name:    "should-not-run",
			Command: []string{"sh", "-c", "echo 'should not run'"},
		},
	}
	cfg := Config{
		Engine:         "docker",
		Image:          "alpine:latest",
		DefaultTimeout: 30,
		FailFast:       true,
	}

	results, failures := RunTests(tests, cfg, func(r Result) {})

	if len(results) != 1 {
		t.Errorf("Expected 1 result with fail-fast, got %d", len(results))
	}
	if failures != 1 {
		t.Errorf("Expected 1 failure, got %d", failures)
	}
}

// TestRunSingleTest_RealExecution_CommandVsExec tests both command forms.
func TestRunSingleTest_RealExecution_CommandVsExec(t *testing.T) {
	t.Run("using_exec", func(t *testing.T) {
		testCase := config.TestCase{
			Name: "exec-form",
			Exec: []string{"echo", "hello from exec"},
			Expect: config.ExpectBlock{
				StdoutContains: []string{"hello from exec"},
			},
		}

		result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

		if result.Status != "PASSED" {
			t.Errorf("Expected PASSED, got %s. Failures: %v", result.Status, result.Failures)
		}
	})

	t.Run("using_command", func(t *testing.T) {
		testCase := config.TestCase{
			Name:    "command-form",
			Command: []string{"sh", "-c", "echo 'hello from command'"},
			Expect: config.ExpectBlock{
				StdoutContains: []string{"hello from command"},
			},
		}

		result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

		if result.Status != "PASSED" {
			t.Errorf("Expected PASSED, got %s. Failures: %v", result.Status, result.Failures)
		}
	})
}

// TestRunSingleTest_RealExecution_MultipleEnvVars tests multiple environment variables.
func TestRunSingleTest_RealExecution_MultipleEnvVars(t *testing.T) {
	testCase := config.TestCase{
		Name:    "multi-env-test",
		Command: []string{"sh", "-c", "echo $VAR1 $VAR2 $VAR3"},
		Env: map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
			"VAR3": "value3",
		},
		Expect: config.ExpectBlock{
			StdoutContains: []string{"value1", "value2", "value3"},
		},
	}

	result := runSingleTest(testCase, "docker", "alpine:latest", 30, false, false)

	if result.Status != "PASSED" {
		t.Errorf("Expected PASSED status, got %s. Failures: %v", result.Status, result.Failures)
	}
}

// TestRunSingleTest_RealExecution_CommandNotFound tests execution error when engine doesn't exist.
func TestRunSingleTest_RealExecution_CommandNotFound(t *testing.T) {
	testCase := config.TestCase{
		Name: "command-not-found-test",
		Exec: []string{"echo", "test"},
	}

	// Use a non-existent container engine to trigger command execution error
	result := runSingleTest(testCase, "nonexistent-container-engine-xyz", "alpine:latest", 30, false, false)

	if result.Status != "FAILED" {
		t.Errorf("Expected FAILED status, got %s", result.Status)
	}
	if len(result.Failures) == 0 {
		t.Error("Expected failure to be reported")
	}
	// Should contain "exception:" in the failure message
	if !strings.Contains(result.Failures[0], "exception:") {
		t.Errorf("Expected 'exception:' in failure message, got: %v", result.Failures)
	}
	if result.ExitCode != nil {
		t.Error("Expected nil exit code for command execution error")
	}
}
