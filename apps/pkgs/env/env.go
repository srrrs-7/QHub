package env

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// GetString returns the value of the environment variable named by the key.
func GetString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New("environment variable not found: " + key)
	}
	return value, nil
}

// GetStringOrDefault returns the value of the environment variable or a default value.
func GetStringOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetInt returns the value of the environment variable named by the key as an integer.
func GetInt(key string) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, errors.New("environment variable not found: " + key)
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}

// GetBool returns the value of the environment variable named by the key as a boolean.
// It accepts "true", "1", "yes" as true values (case-insensitive).
func GetBool(key string) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return false, errors.New("environment variable not found: " + key)
	}

	switch strings.ToLower(value) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, errors.New("invalid boolean value: " + value)
	}
}
