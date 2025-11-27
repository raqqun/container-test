package main

import (
	"os"
	"strconv"
	"strings"
)

// envInt returns an int from env or a fallback.
func envInt(key string, fallback int) int {
    if v, ok := os.LookupEnv(key); ok {
        if parsed, err := strconv.Atoi(v); err == nil {
            return parsed
        }
    }
    return fallback
}

// envBool returns a bool from env or a fallback (true/false/1/0/yes/no/on/off).
func envBool(key string, fallback bool) bool {
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

// getenvDefault returns the environment value or a fallback when unset.
func getenvDefault(key, fallback string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return fallback
}
