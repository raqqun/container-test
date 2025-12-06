package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"container-test-cli/internal/config"
)

// Result captures the outcome of running a single test.
type Result struct {
	Status   string   `json:"status"`
	Name     string   `json:"name"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	ExitCode *int     `json:"exit_code"`
	Failures []string `json:"failures,omitempty"`
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// BuildRunCommand assembles the engine run command with env, workdir, args, and entrypoint.
func BuildRunCommand(engine, image string, cmd []string, workdir string, env map[string]string, runArgs []string, entrypoint *string) []string {
	args := []string{engine, "run", "--rm"}
	args = append(args, runArgs...)
	if entrypoint != nil {
		args = append(args, "--entrypoint", *entrypoint)
	}
	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	if workdir != "" {
		args = append(args, "-w", workdir)
	}
	args = append(args, image)
	args = append(args, cmd...)
	return args
}

// checkContains verifies that output contains all expected strings.
func checkContains(output string, expectedStrings []string, outputName string) []string {
	var failures []string
	for _, needle := range expectedStrings {
		if !strings.Contains(output, needle) {
			failures = append(failures, fmt.Sprintf("%s missing: %q", outputName, needle))
		}
	}
	return failures
}

// checkNotContains verifies that output does not contain forbidden strings.
func checkNotContains(output string, forbiddenStrings []string, outputName string) []string {
	var failures []string
	for _, needle := range forbiddenStrings {
		if strings.Contains(output, needle) {
			failures = append(failures, fmt.Sprintf("%s must not contain: %q", outputName, needle))
		}
	}
	return failures
}

// checkRegex verifies that output matches the given regex pattern.
func checkRegex(output, pattern, outputName string) []string {
	if pattern == "" {
		return nil
	}
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		return []string{fmt.Sprintf("%s does not match regex %q", outputName, pattern)}
	}
	return nil
}

// evalExpectations applies the expectBlock rules to collected outputs.
func evalExpectations(expect config.ExpectBlock, stdout, stderr string, exitCode int) []string {
	var failures []string

	// Check exit code
	expectedExit := config.ExitCodeExpect{Op: "==", Value: 0}
	if expect.ExitCode != nil {
		expectedExit = *expect.ExitCode
	}
	if !expectedExit.SatisfiedBy(exitCode) {
		failures = append(failures, fmt.Sprintf("exit code %d != expected %s", exitCode, expectedExit.String()))
	}

	// Check stdout expectations
	failures = append(failures, checkContains(stdout, expect.StdoutContains, "stdout")...)
	failures = append(failures, checkNotContains(stdout, expect.StdoutNot, "stdout")...)
	failures = append(failures, checkRegex(stdout, expect.StdoutRegex, "stdout")...)

	// Check stderr expectations
	failures = append(failures, checkContains(stderr, expect.StderrContains, "stderr")...)
	failures = append(failures, checkRegex(stderr, expect.StderrRegex, "stderr")...)

	return failures
}

// RunSingleTest executes a single container run and evaluates expectations.
// If dryRun is true, it prints the command without executing it.
func RunSingleTest(testCase config.TestCase, engine, image string, defaultTimeout int, debug, dryRun bool) Result {
	if testCase.Skip {
		return Result{Status: "SKIPPED", Name: firstNonEmpty(testCase.Name, "unnamed")}
	}

	command := testCase.Exec
	if len(command) == 0 {
		command = testCase.Command
	}
	if len(command) == 0 {
		return Result{
			Status:   "FAILED",
			Name:     firstNonEmpty(testCase.Name, "unnamed"),
			Failures: []string{"missing 'exec' or 'command'"},
		}
	}

	runArgs := testCase.RunArgs
	entrypoint := testCase.Entrypoint

	timeout := defaultTimeout
	if testCase.Expect.TimeoutSeconds != nil {
		timeout = *testCase.Expect.TimeoutSeconds
	}
	if testCase.Timeout != nil {
		timeout = *testCase.Timeout
	}

	runCmd := BuildRunCommand(engine, image, command, testCase.Workdir, testCase.Env, runArgs, entrypoint)

	// Handle dry-run mode
	if dryRun {
		fmt.Printf("[dry-run] %s\n", strings.Join(runCmd, " "))
		return Result{
			Status: "DRY-RUN",
			Name:   firstNonEmpty(testCase.Name, "unnamed"),
		}
	}

	if debug {
		fmt.Printf("[debug] running: %s (timeout=%ds)\n", strings.Join(runCmd, " "), timeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, runCmd[0], runCmd[1:]...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()
	exitCode := 0

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{
				Status:   "FAILED",
				Name:     firstNonEmpty(testCase.Name, "unnamed"),
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: nil,
				Failures: []string{fmt.Sprintf("timed out after %ds", timeout)},
			}
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{
				Status:   "FAILED",
				Name:     firstNonEmpty(testCase.Name, "unnamed"),
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: nil,
				Failures: []string{fmt.Sprintf("exception: %v", err)},
			}
		}
	}

	failures := evalExpectations(testCase.Expect, stdout, stderr, exitCode)
	status := "PASSED"
	if len(failures) > 0 {
		status = "FAILED"
	}

	return Result{
		Status:   status,
		Name:     firstNonEmpty(testCase.Name, "unnamed"),
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: &exitCode,
		Failures: failures,
	}
}
