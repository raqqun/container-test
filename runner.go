package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// buildRunCommand assembles the engine run command with env, workdir, args, and entrypoint.
func buildRunCommand(engine, image string, cmd []string, workdir string, env map[string]string, runArgs []string, entrypoint *string) []string {
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

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// evalExpectations applies the expectBlock rules to collected outputs.
func evalExpectations(expect expectBlock, stdout, stderr string, exitCode int) []string {
	failures := []string{}
	expectedExit := exitCodeExpect{Op: "==", Value: 0}
	if expect.ExitCode != nil {
		expectedExit = *expect.ExitCode
	}
	if !expectedExit.satisfiedBy(exitCode) {
		failures = append(failures, fmt.Sprintf("exit code %d != expected %s", exitCode, expectedExit.String()))
	}

	for _, needle := range expect.StdoutContains {
		if !strings.Contains(stdout, needle) {
			failures = append(failures, fmt.Sprintf("stdout missing: %q", needle))
		}
	}
	for _, needle := range expect.StdoutNot {
		if strings.Contains(stdout, needle) {
			failures = append(failures, fmt.Sprintf("stdout must not contain: %q", needle))
		}
	}
	for _, needle := range expect.StderrContains {
		if !strings.Contains(stderr, needle) {
			failures = append(failures, fmt.Sprintf("stderr missing: %q", needle))
		}
	}
	if expect.StdoutRegex != "" {
		re := regexp.MustCompile(expect.StdoutRegex)
		if !re.MatchString(stdout) {
			failures = append(failures, fmt.Sprintf("stdout does not match regex %q", expect.StdoutRegex))
		}
	}
	if expect.StderrRegex != "" {
		re := regexp.MustCompile(expect.StderrRegex)
		if !re.MatchString(stderr) {
			failures = append(failures, fmt.Sprintf("stderr does not match regex %q", expect.StderrRegex))
		}
	}
	return failures
}

// runSingleTest executes a single container run and evaluates expectations.
func runSingleTest(engine, image string, t testCase, defaultTimeout int, debug bool) testResult {
	if t.Skip {
		return testResult{Status: "SKIPPED", Name: firstNonEmpty(t.Name, "unnamed")}
	}

	command := t.Exec
	if len(command) == 0 {
		command = t.Command
	}
	if len(command) == 0 {
		return testResult{
			Status:   "FAILED",
			Name:     firstNonEmpty(t.Name, "unnamed"),
			Failures: []string{"missing 'exec' or 'command'"},
		}
	}

	runArgs := t.RunArgs
	entrypoint := t.Entrypoint

	timeout := defaultTimeout
	if t.Expect.TimeoutSeconds != nil {
		timeout = *t.Expect.TimeoutSeconds
	}
	if t.Timeout != nil {
		timeout = *t.Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	runCmd := buildRunCommand(engine, image, []string(command), t.Workdir, t.Env, runArgs, entrypoint)
	if debug {
		fmt.Printf("[debug] running: %s (timeout=%ds)\n", strings.Join(runCmd, " "), timeout)
	}
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
			return testResult{
				Status:   "FAILED",
				Name:     firstNonEmpty(t.Name, "unnamed"),
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: nil,
				Failures: []string{fmt.Sprintf("timed out after %ds", timeout)},
			}
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return testResult{
				Status:   "FAILED",
				Name:     firstNonEmpty(t.Name, "unnamed"),
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: nil,
				Failures: []string{fmt.Sprintf("exception: %v", err)},
			}
		}
	}

	failures := evalExpectations(t.Expect, stdout, stderr, exitCode)
	status := "PASSED"
	if len(failures) > 0 {
		status = "FAILED"
	}

	return testResult{
		Status:   status,
		Name:     firstNonEmpty(t.Name, "unnamed"),
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: &exitCode,
		Failures: failures,
	}
}
