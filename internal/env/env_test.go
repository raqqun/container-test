package env

import (
	"os"
	"testing"
)

// TestEnvInt_WithValidValue tests EnvInt with a valid integer value.
func TestEnvInt_WithValidValue(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 10)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestEnvInt_WithZeroValue tests EnvInt with zero value.
func TestEnvInt_WithZeroValue(t *testing.T) {
	os.Setenv("TEST_INT", "0")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 10)
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

// TestEnvInt_WithNegativeValue tests EnvInt with a negative integer.
func TestEnvInt_WithNegativeValue(t *testing.T) {
	os.Setenv("TEST_INT", "-100")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 10)
	if result != -100 {
		t.Errorf("Expected -100, got %d", result)
	}
}

// TestEnvInt_WithInvalidValue tests EnvInt with a non-integer value.
func TestEnvInt_WithInvalidValue(t *testing.T) {
	os.Setenv("TEST_INT", "not-a-number")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 99)
	if result != 99 {
		t.Errorf("Expected fallback 99, got %d", result)
	}
}

// TestEnvInt_WithEmptyString tests EnvInt with an empty string.
func TestEnvInt_WithEmptyString(t *testing.T) {
	os.Setenv("TEST_INT", "")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 50)
	if result != 50 {
		t.Errorf("Expected fallback 50, got %d", result)
	}
}

// TestEnvInt_WithMissingKey tests EnvInt when the environment variable is not set.
func TestEnvInt_WithMissingKey(t *testing.T) {
	result := EnvInt("NONEXISTENT_KEY", 123)
	if result != 123 {
		t.Errorf("Expected fallback 123, got %d", result)
	}
}

// TestEnvInt_WithFloatValue tests EnvInt with a float value.
func TestEnvInt_WithFloatValue(t *testing.T) {
	os.Setenv("TEST_INT", "3.14")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 5)
	if result != 5 {
		t.Errorf("Expected fallback 5 for float value, got %d", result)
	}
}

// TestEnvInt_WithWhitespace tests EnvInt with whitespace around the value.
func TestEnvInt_WithWhitespace(t *testing.T) {
	os.Setenv("TEST_INT", "  42  ")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 10)
	// strconv.Atoi doesn't trim whitespace, so this should return fallback
	if result != 10 {
		t.Errorf("Expected fallback 10 for value with whitespace, got %d", result)
	}
}

// TestEnvBool_TrueValues tests EnvBool with all true value formats.
func TestEnvBool_TrueValues(t *testing.T) {
	trueValues := []string{"1", "true", "TRUE", "True", "yes", "YES", "Yes", "on", "ON", "On"}

	for _, val := range trueValues {
		t.Run(val, func(t *testing.T) {
			os.Setenv("TEST_BOOL", val)
			defer os.Unsetenv("TEST_BOOL")

			result := EnvBool("TEST_BOOL", false)
			if !result {
				t.Errorf("Expected true for value '%s', got false", val)
			}
		})
	}
}

// TestEnvBool_FalseValues tests EnvBool with all false value formats.
func TestEnvBool_FalseValues(t *testing.T) {
	falseValues := []string{"0", "false", "FALSE", "False", "no", "NO", "No", "off", "OFF", "Off"}

	for _, val := range falseValues {
		t.Run(val, func(t *testing.T) {
			os.Setenv("TEST_BOOL", val)
			defer os.Unsetenv("TEST_BOOL")

			result := EnvBool("TEST_BOOL", true)
			if result {
				t.Errorf("Expected false for value '%s', got true", val)
			}
		})
	}
}

// TestEnvBool_InvalidValueWithFalseFallback tests EnvBool with invalid value and false fallback.
func TestEnvBool_InvalidValueWithFalseFallback(t *testing.T) {
	os.Setenv("TEST_BOOL", "invalid")
	defer os.Unsetenv("TEST_BOOL")

	result := EnvBool("TEST_BOOL", false)
	if result {
		t.Error("Expected false fallback for invalid value, got true")
	}
}

// TestEnvBool_InvalidValueWithTrueFallback tests EnvBool with invalid value and true fallback.
func TestEnvBool_InvalidValueWithTrueFallback(t *testing.T) {
	os.Setenv("TEST_BOOL", "invalid")
	defer os.Unsetenv("TEST_BOOL")

	result := EnvBool("TEST_BOOL", true)
	if !result {
		t.Error("Expected true fallback for invalid value, got false")
	}
}

// TestEnvBool_EmptyStringWithFalseFallback tests EnvBool with empty string and false fallback.
func TestEnvBool_EmptyStringWithFalseFallback(t *testing.T) {
	os.Setenv("TEST_BOOL", "")
	defer os.Unsetenv("TEST_BOOL")

	result := EnvBool("TEST_BOOL", false)
	if result {
		t.Error("Expected false fallback for empty string, got true")
	}
}

// TestEnvBool_EmptyStringWithTrueFallback tests EnvBool with empty string and true fallback.
func TestEnvBool_EmptyStringWithTrueFallback(t *testing.T) {
	os.Setenv("TEST_BOOL", "")
	defer os.Unsetenv("TEST_BOOL")

	result := EnvBool("TEST_BOOL", true)
	if !result {
		t.Error("Expected true fallback for empty string, got false")
	}
}

// TestEnvBool_MissingKeyWithFalseFallback tests EnvBool when the key is not set and fallback is false.
func TestEnvBool_MissingKeyWithFalseFallback(t *testing.T) {
	result := EnvBool("NONEXISTENT_BOOL", false)
	if result {
		t.Error("Expected false fallback for missing key, got true")
	}
}

// TestEnvBool_MissingKeyWithTrueFallback tests EnvBool when the key is not set and fallback is true.
func TestEnvBool_MissingKeyWithTrueFallback(t *testing.T) {
	result := EnvBool("NONEXISTENT_BOOL", true)
	if !result {
		t.Error("Expected true fallback for missing key, got false")
	}
}

// TestEnvBool_WithWhitespace tests EnvBool with whitespace around the value.
func TestEnvBool_WithWhitespace(t *testing.T) {
	os.Setenv("TEST_BOOL", "  true  ")
	defer os.Unsetenv("TEST_BOOL")

	result := EnvBool("TEST_BOOL", false)
	// The implementation uses strings.ToLower which doesn't trim, so "  true  " won't match
	// and should return fallback
	if result {
		t.Error("Expected false fallback for value with whitespace, got true")
	}
}

// TestEnvBool_MixedCase tests EnvBool with mixed case values.
func TestEnvBool_MixedCase(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"TrUe", "TrUe", true},
		{"YeS", "YeS", true},
		{"FaLsE", "FaLsE", false},
		{"nO", "nO", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tt.value)
			defer os.Unsetenv("TEST_BOOL")

			result := EnvBool("TEST_BOOL", false)
			if result != tt.expected {
				t.Errorf("Expected %v for value '%s', got %v", tt.expected, tt.value, result)
			}
		})
	}
}

// TestEnvDefault_WithValue tests EnvDefault with a set value.
func TestEnvDefault_WithValue(t *testing.T) {
	os.Setenv("TEST_STRING", "hello")
	defer os.Unsetenv("TEST_STRING")

	result := EnvDefault("TEST_STRING", "default")
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

// TestEnvDefault_WithEmptyString tests EnvDefault with an empty string value.
func TestEnvDefault_WithEmptyString(t *testing.T) {
	os.Setenv("TEST_STRING", "")
	defer os.Unsetenv("TEST_STRING")

	result := EnvDefault("TEST_STRING", "default")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

// TestEnvDefault_WithMissingKey tests EnvDefault when the key is not set.
func TestEnvDefault_WithMissingKey(t *testing.T) {
	result := EnvDefault("NONEXISTENT_STRING", "fallback")
	if result != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", result)
	}
}

// TestEnvDefault_WithWhitespace tests EnvDefault with whitespace values.
func TestEnvDefault_WithWhitespace(t *testing.T) {
	os.Setenv("TEST_STRING", "  spaces  ")
	defer os.Unsetenv("TEST_STRING")

	result := EnvDefault("TEST_STRING", "default")
	if result != "  spaces  " {
		t.Errorf("Expected '  spaces  ', got '%s'", result)
	}
}

// TestEnvDefault_WithSpecialCharacters tests EnvDefault with special characters.
func TestEnvDefault_WithSpecialCharacters(t *testing.T) {
	specialValue := "!@#$%^&*()_+-={}[]|:;<>?,./~`"
	os.Setenv("TEST_STRING", specialValue)
	defer os.Unsetenv("TEST_STRING")

	result := EnvDefault("TEST_STRING", "default")
	if result != specialValue {
		t.Errorf("Expected '%s', got '%s'", specialValue, result)
	}
}

// TestEnvDefault_WithMultilineValue tests EnvDefault with multiline values.
func TestEnvDefault_WithMultilineValue(t *testing.T) {
	multiline := "line1\nline2\nline3"
	os.Setenv("TEST_STRING", multiline)
	defer os.Unsetenv("TEST_STRING")

	result := EnvDefault("TEST_STRING", "default")
	if result != multiline {
		t.Errorf("Expected multiline value, got '%s'", result)
	}
}

// TestEnvDefault_WithUnicodeCharacters tests EnvDefault with unicode characters.
func TestEnvDefault_WithUnicodeCharacters(t *testing.T) {
	unicode := "Hello 世界 🌍"
	os.Setenv("TEST_STRING", unicode)
	defer os.Unsetenv("TEST_STRING")

	result := EnvDefault("TEST_STRING", "default")
	if result != unicode {
		t.Errorf("Expected '%s', got '%s'", unicode, result)
	}
}

// TestEnvDefault_EmptyFallback tests EnvDefault with empty fallback value.
func TestEnvDefault_EmptyFallback(t *testing.T) {
	result := EnvDefault("NONEXISTENT_STRING", "")
	if result != "" {
		t.Errorf("Expected empty string fallback, got '%s'", result)
	}
}

// TestEnvInt_LargeNumbers tests EnvInt with large integer values.
func TestEnvInt_LargeNumbers(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected int
	}{
		{"max_int32", "2147483647", 2147483647},
		{"min_int32", "-2147483648", -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_INT", tt.value)
			defer os.Unsetenv("TEST_INT")

			result := EnvInt("TEST_INT", 0)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestEnvInt_WithLeadingZeros tests EnvInt with leading zeros.
func TestEnvInt_WithLeadingZeros(t *testing.T) {
	os.Setenv("TEST_INT", "00042")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 0)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestEnvInt_WithPlusSign tests EnvInt with explicit plus sign.
func TestEnvInt_WithPlusSign(t *testing.T) {
	os.Setenv("TEST_INT", "+42")
	defer os.Unsetenv("TEST_INT")

	result := EnvInt("TEST_INT", 0)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// TestMultipleFunctions_Together tests using multiple env functions in sequence.
func TestMultipleFunctions_Together(t *testing.T) {
	// Set up multiple environment variables
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_NAME", "test-app")

	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_DEBUG")
		os.Unsetenv("APP_NAME")
	}()

	// Test all three functions
	port := EnvInt("APP_PORT", 3000)
	debug := EnvBool("APP_DEBUG", false)
	name := EnvDefault("APP_NAME", "default-app")

	if port != 8080 {
		t.Errorf("Expected port 8080, got %d", port)
	}
	if !debug {
		t.Error("Expected debug to be true")
	}
	if name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", name)
	}
}

// TestEnvFunctions_Isolation tests that functions don't interfere with each other.
func TestEnvFunctions_Isolation(t *testing.T) {
	os.Setenv("SHARED_KEY", "value")
	defer os.Unsetenv("SHARED_KEY")

	// Each function should handle the same key differently
	intResult := EnvInt("SHARED_KEY", 100)
	boolResult := EnvBool("SHARED_KEY", false)
	stringResult := EnvDefault("SHARED_KEY", "default")

	// EnvInt should fail to parse "value" as int and return fallback
	if intResult != 100 {
		t.Errorf("Expected int fallback 100, got %d", intResult)
	}

	// EnvBool should fail to parse "value" as bool and return fallback
	if boolResult {
		t.Error("Expected bool fallback false, got true")
	}

	// EnvDefault should return the actual value
	if stringResult != "value" {
		t.Errorf("Expected 'value', got '%s'", stringResult)
	}
}
