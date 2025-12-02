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
		out := make([]string, 0, len(value.Content))
		for _, n := range value.Content {
			out = append(out, n.Value)
		}
		*s = out
		return nil
	default:
		return fmt.Errorf("value must be string or list")
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
		for _, n := range value.Content {
			result = append(result, n.Value)
		}
		*c = result
		return nil
	default:
		return fmt.Errorf("command must be string or list")
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

// UnmarshalYAML accepts an int or a comparison expression string (==, !=, >=, <=, >, <).
func (e *ExitCodeExpect) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("exit_code must be an int or comparison expression")
	}

	if v, err := strconv.Atoi(value.Value); err == nil {
		e.Op = "=="
		e.Value = v
		e.RawInt = &v
		return nil
	}

	expr := strings.TrimSpace(value.Value)
	op := "=="
	rest := expr
	for _, candidate := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if strings.HasPrefix(expr, candidate) {
			op = candidate
			rest = strings.TrimSpace(expr[len(candidate):])
			break
		}
	}

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
		var tests []TestCase
		if err := value.Decode(&tests); err != nil {
			return err
		}
		*tl = tests
		return nil
	}
	if value.Kind == yaml.MappingNode {
		var container struct {
			Tests []TestCase `yaml:"tests"`
		}
		if err := value.Decode(&container); err != nil {
			return err
		}
		*tl = container.Tests
		return nil
	}
	return fmt.Errorf("config must be a list or contain a 'tests' list")
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
