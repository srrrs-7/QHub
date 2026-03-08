package env

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"utils/types"
)

// GetString returns the value of the environment variable named by the key.
func GetString(key string) types.Result[string, error] {
	value := os.Getenv(key)
	if value == "" {
		return types.Err[string](errors.New("environment variable not found: " + key))
	}
	return types.Ok[string, error](value)
}

// GetInt returns the value of the environment variable named by the key as an integer.
func GetInt(key string) types.Result[int, error] {
	value := os.Getenv(key)
	if value == "" {
		return types.Err[int](errors.New("environment variable not found: " + key))
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return types.Err[int](err)
	}
	return types.Ok[int, error](result)
}

// GetBool returns the value of the environment variable named by the key as a boolean.
// It accepts "true", "1", "yes" as true values (case-insensitive).
func GetBool(key string) types.Result[bool, error] {
	value := os.Getenv(key)
	if value == "" {
		return types.Err[bool](errors.New("environment variable not found: " + key))
	}

	switch strings.ToLower(value) {
	case "true", "1", "yes":
		return types.Ok[bool, error](true)
	case "false", "0", "no":
		return types.Ok[bool, error](false)
	default:
		return types.Err[bool](errors.New("invalid boolean value: " + value))
	}
}
