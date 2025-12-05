package env

import (
	"os"
	"strconv"
	"strings"
)

// EnvInt returns an int from env or a fallback.
func EnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

// EnvBool returns a bool from env or a fallback (true/false/1/0/yes/no/on/off).
func EnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		default:
			return fallback
		}
	}
	return fallback
}

// EnvDefault returns the environment value or a fallback when unset.
func EnvDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
