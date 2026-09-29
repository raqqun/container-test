package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestStringOrSlice_UnmarshalYAML_SingleString tests unmarshaling a single string.
func TestStringOrSlice_UnmarshalYAML_SingleString(t *testing.T) {
	yamlData := `value: "single string"`

	var result struct {
		Value StringOrSlice `yaml:"value"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.Value) != 1 {
		t.Errorf("Expected 1 element, got %d", len(result.Value))
	}
	if result.Value[0] != "single string" {
		t.Errorf("Expected 'single string', got '%s'", result.Value[0])
	}
}

// TestStringOrSlice_UnmarshalYAML_List tests unmarshaling a list of strings.
func TestStringOrSlice_UnmarshalYAML_List(t *testing.T) {
	yamlData := `value:
  - first
  - second
  - third`

	var result struct {
		Value StringOrSlice `yaml:"value"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"first", "second", "third"}
	if len(result.Value) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(result.Value))
	}

	for i, exp := range expected {
		if result.Value[i] != exp {
			t.Errorf("Element %d: expected '%s', got '%s'", i, exp, result.Value[i])
		}
	}
}

// TestStringOrSlice_UnmarshalYAML_EmptyList tests unmarshaling an empty list.
func TestStringOrSlice_UnmarshalYAML_EmptyList(t *testing.T) {
	yamlData := `value: []`

	var result struct {
		Value StringOrSlice `yaml:"value"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.Value) != 0 {
		t.Errorf("Expected empty list, got %d elements", len(result.Value))
	}
}

// TestStringOrSlice_UnmarshalYAML_Invalid tests unmarshaling invalid types.
func TestStringOrSlice_UnmarshalYAML_Invalid(t *testing.T) {
	yamlData := `value:
  key: nested`

	var result struct {
		Value StringOrSlice `yaml:"value"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err == nil {
		t.Fatal("Expected error for invalid type, got nil")
	}

	if !strings.Contains(err.Error(), "must be a string or a list") {
		t.Errorf("Expected error about string or list, got: %v", err)
	}
}

// TestCommandValue_UnmarshalYAML_String tests unmarshaling a command string.
func TestCommandValue_UnmarshalYAML_String(t *testing.T) {
	yamlData := `cmd: "echo hello"`

	var result struct {
		Cmd CommandValue `yaml:"cmd"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"sh", "-c", "echo hello"}
	if len(result.Cmd) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(result.Cmd))
	}

	for i, exp := range expected {
		if result.Cmd[i] != exp {
			t.Errorf("Element %d: expected '%s', got '%s'", i, exp, result.Cmd[i])
		}
	}
}

// TestCommandValue_UnmarshalYAML_List tests unmarshaling a command list.
func TestCommandValue_UnmarshalYAML_List(t *testing.T) {
	yamlData := `cmd:
  - /bin/echo
  - hello
  - world`

	var result struct {
		Cmd CommandValue `yaml:"cmd"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"/bin/echo", "hello", "world"}
	if len(result.Cmd) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(result.Cmd))
	}

	for i, exp := range expected {
		if result.Cmd[i] != exp {
			t.Errorf("Element %d: expected '%s', got '%s'", i, exp, result.Cmd[i])
		}
	}
}

// TestCommandValue_UnmarshalYAML_Invalid tests unmarshaling invalid command types.
func TestCommandValue_UnmarshalYAML_Invalid(t *testing.T) {
	yamlData := `cmd:
  key: value`

	var result struct {
		Cmd CommandValue `yaml:"cmd"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err == nil {
		t.Fatal("Expected error for invalid type, got nil")
	}

	if !strings.Contains(err.Error(), "command must be a string or a list") {
		t.Errorf("Expected error about command type, got: %v", err)
	}
}

// TestExitCodeExpect_UnmarshalYAML_PlainInteger tests unmarshaling a plain integer.
func TestExitCodeExpect_UnmarshalYAML_PlainInteger(t *testing.T) {
	yamlData := `exit_code: 0`

	var result struct {
		ExitCode ExitCodeExpect `yaml:"exit_code"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.ExitCode.Op != "==" {
		t.Errorf("Expected operator '==', got '%s'", result.ExitCode.Op)
	}
	if result.ExitCode.Value != 0 {
		t.Errorf("Expected value 0, got %d", result.ExitCode.Value)
	}
	if result.ExitCode.RawInt == nil || *result.ExitCode.RawInt != 0 {
		t.Error("Expected RawInt to be set to 0")
	}
}

// TestExitCodeExpect_UnmarshalYAML_Operators tests all comparison operators.
func TestExitCodeExpect_UnmarshalYAML_Operators(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		wantOp   string
		wantVal  int
	}{
		{"equals", `exit_code: "==0"`, "==", 0},
		{"not_equals", `exit_code: "!=1"`, "!=", 1},
		{"greater_than_or_equal", `exit_code: ">=0"`, ">=", 0},
		{"less_than_or_equal", `exit_code: "<=2"`, "<=", 2},
		{"greater_than", `exit_code: ">0"`, ">", 0},
		{"less_than", `exit_code: "<10"`, "<", 10},
		{"negative_value", `exit_code: "==-1"`, "==", -1},
		{"with_spaces", `exit_code: ">=  5"`, ">=", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				ExitCode ExitCodeExpect `yaml:"exit_code"`
			}

			err := yaml.Unmarshal([]byte(tt.yamlData), &result)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.ExitCode.Op != tt.wantOp {
				t.Errorf("Expected operator '%s', got '%s'", tt.wantOp, result.ExitCode.Op)
			}
			if result.ExitCode.Value != tt.wantVal {
				t.Errorf("Expected value %d, got %d", tt.wantVal, result.ExitCode.Value)
			}
			if result.ExitCode.RawInt != nil {
				t.Error("Expected RawInt to be nil for expression form")
			}
		})
	}
}

// TestExitCodeExpect_UnmarshalYAML_Invalid tests unmarshaling invalid exit codes.
func TestExitCodeExpect_UnmarshalYAML_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		errMsg   string
	}{
		{"not_scalar", `exit_code: []`, "must be an integer or a comparison expression"},
		{"invalid_expression", `exit_code: "==abc"`, "must end with an integer"},
		{"no_number", `exit_code: ">=  "`, "must end with an integer"},
		{"invalid_text", `exit_code: "invalid"`, "must end with an integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				ExitCode ExitCodeExpect `yaml:"exit_code"`
			}

			err := yaml.Unmarshal([]byte(tt.yamlData), &result)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing '%s', got: %v", tt.errMsg, err)
			}
		})
	}
}

// TestExitCodeExpect_SatisfiedBy tests exit code comparison logic.
func TestExitCodeExpect_SatisfiedBy(t *testing.T) {
	tests := []struct {
		name     string
		expect   ExitCodeExpect
		actual   int
		expected bool
	}{
		{"equals_match", ExitCodeExpect{Op: "==", Value: 0}, 0, true},
		{"equals_no_match", ExitCodeExpect{Op: "==", Value: 0}, 1, false},
		{"not_equals_match", ExitCodeExpect{Op: "!=", Value: 0}, 1, true},
		{"not_equals_no_match", ExitCodeExpect{Op: "!=", Value: 0}, 0, false},
		{"greater_than_match", ExitCodeExpect{Op: ">", Value: 5}, 10, true},
		{"greater_than_no_match", ExitCodeExpect{Op: ">", Value: 5}, 5, false},
		{"greater_than_or_equal_match", ExitCodeExpect{Op: ">=", Value: 5}, 5, true},
		{"greater_than_or_equal_no_match", ExitCodeExpect{Op: ">=", Value: 5}, 4, false},
		{"less_than_match", ExitCodeExpect{Op: "<", Value: 5}, 2, true},
		{"less_than_no_match", ExitCodeExpect{Op: "<", Value: 5}, 5, false},
		{"less_than_or_equal_match", ExitCodeExpect{Op: "<=", Value: 5}, 5, true},
		{"less_than_or_equal_no_match", ExitCodeExpect{Op: "<=", Value: 5}, 6, false},
		{"invalid_operator", ExitCodeExpect{Op: "??", Value: 0}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expect.SatisfiedBy(tt.actual)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestExitCodeExpect_String tests string representation.
func TestExitCodeExpect_String(t *testing.T) {
	tests := []struct {
		name     string
		expect   ExitCodeExpect
		expected string
	}{
		{"equals", ExitCodeExpect{Op: "==", Value: 0}, "==0"},
		{"not_equals", ExitCodeExpect{Op: "!=", Value: 1}, "!=1"},
		{"greater_than", ExitCodeExpect{Op: ">", Value: 5}, ">5"},
		{"less_than", ExitCodeExpect{Op: "<", Value: 10}, "<10"},
		{"negative", ExitCodeExpect{Op: "==", Value: -1}, "==-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expect.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestParseOperator tests operator parsing.
func TestParseOperator(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantOp   string
		wantRest string
	}{
		{"equals", "==5", "==", "5"},
		{"not_equals", "!=10", "!=", "10"},
		{"greater_than_or_equal", ">=0", ">=", "0"},
		{"less_than_or_equal", "<=100", "<=", "100"},
		{"greater_than", ">5", ">", "5"},
		{"less_than", "<10", "<", "10"},
		{"no_operator", "42", "==", "42"},
		{"with_spaces", ">=  10", ">=", "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, rest := ExitCodeExpect{}.parseOperator(tt.expr)
			if op != tt.wantOp {
				t.Errorf("Expected operator '%s', got '%s'", tt.wantOp, op)
			}
			if rest != tt.wantRest {
				t.Errorf("Expected rest '%s', got '%s'", tt.wantRest, rest)
			}
		})
	}
}

// TestTestList_UnmarshalYAML_RootArray tests unmarshaling a root array.
func TestTestList_UnmarshalYAML_RootArray(t *testing.T) {
	yamlData := `- name: test1
  exec: ["echo", "hello"]
- name: test2
  command: "echo world"`

	var tests TestList
	err := yaml.Unmarshal([]byte(yamlData), &tests)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tests) != 2 {
		t.Fatalf("Expected 2 tests, got %d", len(tests))
	}

	if tests[0].Name != "test1" {
		t.Errorf("Expected name 'test1', got '%s'", tests[0].Name)
	}
	if tests[1].Name != "test2" {
		t.Errorf("Expected name 'test2', got '%s'", tests[1].Name)
	}
}

// TestTestList_UnmarshalYAML_TestsKey tests unmarshaling with {tests: []} format.
func TestTestList_UnmarshalYAML_TestsKey(t *testing.T) {
	yamlData := `tests:
  - name: test1
    exec: ["echo", "hello"]
  - name: test2
    command: "echo world"`

	var tests TestList
	err := yaml.Unmarshal([]byte(yamlData), &tests)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tests) != 2 {
		t.Fatalf("Expected 2 tests, got %d", len(tests))
	}

	if tests[0].Name != "test1" {
		t.Errorf("Expected name 'test1', got '%s'", tests[0].Name)
	}
	if tests[1].Name != "test2" {
		t.Errorf("Expected name 'test2', got '%s'", tests[1].Name)
	}
}

// TestTestList_UnmarshalYAML_Invalid tests unmarshaling invalid formats.
func TestTestList_UnmarshalYAML_Invalid(t *testing.T) {
	yamlData := `"just a string"`

	var tests TestList
	err := yaml.Unmarshal([]byte(yamlData), &tests)
	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}

	if !strings.Contains(err.Error(), "must be a list of tests or a map containing a 'tests' key") {
		t.Errorf("Expected error about format, got: %v", err)
	}
}

// TestTestCase_ResolveName tests name resolution with and without name field.
func TestTestCase_ResolveName(t *testing.T) {
	tests := []struct {
		name     string
		testCase TestCase
		index    int
		expected string
	}{
		{"with_name", TestCase{Name: "my-test"}, 0, "my-test"},
		{"without_name_first", TestCase{Name: ""}, 0, "test-1"},
		{"without_name_second", TestCase{Name: ""}, 1, "test-2"},
		{"without_name_tenth", TestCase{Name: ""}, 9, "test-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testCase.ResolveName(tt.index)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestLoadTests_ValidFile tests loading a valid test file.
func TestLoadTests_ValidFile(t *testing.T) {
	content := `- name: test1
  exec: ["echo", "hello"]
  expect:
    exit_code: 0
- name: test2
  command: "echo world"
  skip: true`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tests.yaml")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tests, err := LoadTests(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tests) != 2 {
		t.Fatalf("Expected 2 tests, got %d", len(tests))
	}

	if tests[0].Name != "test1" {
		t.Errorf("Expected name 'test1', got '%s'", tests[0].Name)
	}
	if tests[1].Skip != true {
		t.Error("Expected test2 to be skipped")
	}
}

// TestLoadTests_MissingFile tests loading a missing file.
func TestLoadTests_MissingFile(t *testing.T) {
	_, err := LoadTests("/nonexistent/path/tests.yaml")
	if err == nil {
		t.Fatal("Expected error for missing file, got nil")
	}
}

// TestLoadTests_InvalidYAML tests loading a file with invalid YAML.
func TestLoadTests_InvalidYAML(t *testing.T) {
	content := `- name: test1
  exec: [broken yaml`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	_, err = LoadTests(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

// TestLoadTests_InvalidRegexes tests that invalid regex patterns are rejected
// at load time with the offending test name and field (V-02).
func TestLoadTests_InvalidRegexes(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		errParts []string
	}{
		{
			name: "stdout_regex",
			content: `tests:
  - name: bad-regex
    command: "echo hi"
    expect:
      stdout_regex: "("`,
			errParts: []string{`test "bad-regex"`, `stdout_regex`, `invalid`},
		},
		{
			name: "stderr_regex",
			content: `tests:
  - name: bad-stderr
    command: "echo hi"
    expect:
      stderr_regex: "["`,
			errParts: []string{`test "bad-stderr"`, `stderr_regex`, `invalid`},
		},
		{
			name: "default_name",
			content: `tests:
  - exec: ["echo"]
    expect:
      stdout_regex: "("`,
			errParts: []string{`test "test-1"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "badregex.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}

			tests, err := LoadTests(tmpFile)
			if err == nil {
				t.Fatalf("Expected error for invalid regex, got tests: %v", tests)
			}
			for _, part := range tt.errParts {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("Expected error to contain %q, got: %v", part, err)
				}
			}
		})
	}
}

// TestLoadTests_ValidRegexesAccepted tests that valid regex patterns load fine.
func TestLoadTests_ValidRegexesAccepted(t *testing.T) {
	content := `tests:
  - name: ok
    exec: ["echo", "hello"]
    expect:
      stdout_regex: "^curl ?[0-9]+\\.[0-9]+\\.[0-9]+"
      stderr_regex: ""`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "okregex.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tests, err := LoadTests(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tests) != 1 || tests[0].Name != "ok" {
		t.Errorf("Expected 1 test named 'ok', got: %v", tests)
	}
}

// TestTestCase_ComplexStructure tests unmarshaling a complete test case.
func TestTestCase_ComplexStructure(t *testing.T) {
	yamlData := `name: complex-test
command: "echo test"
skip: false
workdir: /app
env:
  KEY1: value1
  KEY2: value2
run_args:
  - --network=host
  - --rm
entrypoint: "/bin/sh"
timeout_seconds: 60
expect:
  exit_code: ">=0"
  stdout_contains:
    - "test"
    - "output"
  stdout_not_contains: "error"
  stdout_regex: "^test.*"
  stderr_contains: "warning"
  stderr_regex: "warn.*"
  timeout_seconds: 30`

	var testCase TestCase
	err := yaml.Unmarshal([]byte(yamlData), &testCase)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if testCase.Name != "complex-test" {
		t.Errorf("Expected name 'complex-test', got '%s'", testCase.Name)
	}
	if testCase.Workdir != "/app" {
		t.Errorf("Expected workdir '/app', got '%s'", testCase.Workdir)
	}
	if len(testCase.Env) != 2 {
		t.Errorf("Expected 2 env vars, got %d", len(testCase.Env))
	}
	if testCase.Env["KEY1"] != "value1" {
		t.Errorf("Expected KEY1='value1', got '%s'", testCase.Env["KEY1"])
	}
	if len(testCase.RunArgs) != 2 {
		t.Errorf("Expected 2 run args, got %d", len(testCase.RunArgs))
	}
	if testCase.Entrypoint == nil || *testCase.Entrypoint != "/bin/sh" {
		t.Error("Expected entrypoint '/bin/sh'")
	}
	if testCase.Timeout == nil || *testCase.Timeout != 60 {
		t.Error("Expected timeout 60")
	}
	if testCase.Expect.ExitCode == nil {
		t.Fatal("Expected exit_code to be set")
	}
	if testCase.Expect.ExitCode.Op != ">=" || testCase.Expect.ExitCode.Value != 0 {
		t.Error("Expected exit_code >=0")
	}
	if len(testCase.Expect.StdoutContains) != 2 {
		t.Errorf("Expected 2 stdout_contains, got %d", len(testCase.Expect.StdoutContains))
	}
	if len(testCase.Expect.StdoutNot) != 1 {
		t.Errorf("Expected 1 stdout_not_contains, got %d", len(testCase.Expect.StdoutNot))
	}
	if testCase.Expect.StdoutRegex != "^test.*" {
		t.Errorf("Expected stdout_regex '^test.*', got '%s'", testCase.Expect.StdoutRegex)
	}
	if testCase.Expect.TimeoutSeconds == nil || *testCase.Expect.TimeoutSeconds != 30 {
		t.Error("Expected expect.timeout_seconds 30")
	}
}

// TestExpectBlock_Optional tests that ExpectBlock fields are optional.
func TestExpectBlock_Optional(t *testing.T) {
	yamlData := `name: minimal-test
exec: ["echo", "hello"]`

	var testCase TestCase
	err := yaml.Unmarshal([]byte(yamlData), &testCase)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if testCase.Expect.ExitCode != nil {
		t.Error("Expected exit_code to be nil")
	}
	if len(testCase.Expect.StdoutContains) != 0 {
		t.Error("Expected empty stdout_contains")
	}
	if len(testCase.Expect.StdoutNot) != 0 {
		t.Error("Expected empty stdout_not_contains")
	}
	if testCase.Expect.StdoutRegex != "" {
		t.Error("Expected empty stdout_regex")
	}
	if testCase.Expect.TimeoutSeconds != nil {
		t.Error("Expected timeout_seconds to be nil")
	}
}

// TestTestCase_BothExecAndCommand tests that both exec and command can be present.
func TestTestCase_BothExecAndCommand(t *testing.T) {
	yamlData := `name: test
exec: ["echo", "from-exec"]
command: "echo from-command"`

	var testCase TestCase
	err := yaml.Unmarshal([]byte(yamlData), &testCase)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Both should be populated
	if len(testCase.Exec) == 0 {
		t.Error("Expected exec to be populated")
	}
	if len(testCase.Command) == 0 {
		t.Error("Expected command to be populated")
	}
}

// TestLoadTests_EmptyFile tests loading an empty file.
func TestLoadTests_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.yaml")

	err := os.WriteFile(tmpFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tests, err := LoadTests(tmpFile)
	// Empty file should not produce an error but return empty list or error depending on YAML parser
	// In this case, it returns an error because the YAML is invalid (empty doc)
	if err == nil {
		// If no error, should be empty list
		if len(tests) != 0 {
			t.Errorf("Expected 0 tests for empty file, got %d", len(tests))
		}
	}
	// Both behaviors (error or empty list) are acceptable for empty file
}

// TestLoadTests_WithTestsKeyEmpty tests loading a file with empty tests array.
func TestLoadTests_WithTestsKeyEmpty(t *testing.T) {
	content := `tests: []`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty-tests.yaml")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tests, err := LoadTests(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tests) != 0 {
		t.Errorf("Expected 0 tests, got %d", len(tests))
	}
}
