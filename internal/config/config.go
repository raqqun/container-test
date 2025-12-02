package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// StringOrSlice allows YAML fields to be single string or list.
type StringOrSlice []string

// UnmarshalYAML allows either a single string or a list of strings in YAML.
func (s *StringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*s = []string{value.Value}
		return nil
	case yaml.SequenceNode:
		result := make([]string, 0, len(value.Content))
		for _, node := range value.Content {
			result = append(result, node.Value)
		}
		*s = result
		return nil
	default:
		return fmt.Errorf("value must be a string or a list of strings")
	}
}

// CommandValue holds a normalized argv form of a command.
type CommandValue []string

// UnmarshalYAML accepts either a shell string (wrapped with sh -c) or argv list.
func (c *CommandValue) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*c = []string{"sh", "-c", value.Value}
		return nil
	case yaml.SequenceNode:
		result := make([]string, 0, len(value.Content))
		for _, node := range value.Content {
			result = append(result, node.Value)
		}
		*c = result
		return nil
	default:
		return fmt.Errorf("command must be a string or a list of strings")
	}
}

// ExpectBlock defines assertions for a test's stdout/stderr/exit code.
type ExpectBlock struct {
	ExitCode       *ExitCodeExpect `yaml:"exit_code" json:"exit_code,omitempty"`
	StdoutContains StringOrSlice   `yaml:"stdout_contains" json:"stdout_contains,omitempty"`
	StdoutNot      StringOrSlice   `yaml:"stdout_not_contains" json:"stdout_not_contains,omitempty"`
	StderrContains StringOrSlice   `yaml:"stderr_contains" json:"stderr_contains,omitempty"`
	StdoutRegex    string          `yaml:"stdout_regex" json:"stdout_regex,omitempty"`
	StderrRegex    string          `yaml:"stderr_regex" json:"stderr_regex,omitempty"`
	TimeoutSeconds *int            `yaml:"timeout_seconds" json:"timeout_seconds,omitempty"`
}

// ExitCodeExpect holds a parsed exit-code expression like ==0, >=1, !=0, <2.
type ExitCodeExpect struct {
	Op     string `json:"op"`
	Value  int    `json:"value"`
	RawInt *int   `json:"raw,omitempty"` // preserve original int for reporting/compatibility
}

// String returns a string representation of the exit code expectation.
func (e ExitCodeExpect) String() string {
	return fmt.Sprintf("%s%d", e.Op, e.Value)
}

// SatisfiedBy checks whether the actual exit code matches the expectation.
func (e ExitCodeExpect) SatisfiedBy(actual int) bool {
	switch e.Op {
	case "==":
		return actual == e.Value
	case "!=":
		return actual != e.Value
	case ">=":
		return actual >= e.Value
	case "<=":
		return actual <= e.Value
	case ">":
		return actual > e.Value
	case "<":
		return actual < e.Value
	default:
		return false
	}
}

// parseOperator extracts the comparison operator from an expression.
// Returns the operator and the remaining string after the operator.
func parseOperator(expr string) (string, string) {
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}
	for _, op := range operators {
		if strings.HasPrefix(expr, op) {
			rest := strings.TrimSpace(expr[len(op):])
			return op, rest
		}
	}
	return "==", expr
}

// UnmarshalYAML accepts an int or a comparison expression string (==, !=, >=, <=, >, <).
func (e *ExitCodeExpect) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("exit_code must be an integer or a comparison expression")
	}

	// Try parsing as a plain integer
	if v, err := strconv.Atoi(value.Value); err == nil {
		e.Op = "=="
		e.Value = v
		e.RawInt = &v
		return nil
	}

	// Parse as a comparison expression
	expr := strings.TrimSpace(value.Value)
	op, rest := parseOperator(expr)

	v, err := strconv.Atoi(rest)
	if err != nil {
		return fmt.Errorf("exit_code expression must end with an integer: %w", err)
	}

	e.Op = op
	e.Value = v
	return nil
}

// TestCase models a single test definition from YAML.
type TestCase struct {
	Name       string            `yaml:"name" json:"name"`
	Exec       CommandValue      `yaml:"exec" json:"exec"`
	Command    CommandValue      `yaml:"command" json:"command"`
	Skip       bool              `yaml:"skip" json:"skip"`
	Workdir    string            `yaml:"workdir" json:"workdir,omitempty"`
	Env        map[string]string `yaml:"env" json:"env,omitempty"`
	Expect     ExpectBlock       `yaml:"expect" json:"expect,omitempty"`
	RunArgs    []string          `yaml:"run_args" json:"run_args,omitempty"`
	Entrypoint *string           `yaml:"entrypoint" json:"entrypoint,omitempty"`
	Timeout    *int              `yaml:"timeout_seconds" json:"timeout_seconds,omitempty"`
}

// TestList holds all parsed test cases.
type TestList []TestCase

// UnmarshalYAML supports a root sequence or {tests: []}.
func (tl *TestList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		result := make([]TestCase, 0, len(value.Content))
		if err := value.Decode(&result); err != nil {
			return err
		}
		*tl = result
		return nil
	}
	if value.Kind == yaml.MappingNode {
		var wrapper struct {
			Tests []TestCase `yaml:"tests"`
		}
		if err := value.Decode(&wrapper); err != nil {
			return err
		}
		*tl = wrapper.Tests
		return nil
	}
	return fmt.Errorf("config must be a list of tests or a map containing a 'tests' key")
}

// LoadTests reads and parses the YAML test definitions.
func LoadTests(path string) (TestList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tests TestList
	if err := yaml.Unmarshal(data, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}
