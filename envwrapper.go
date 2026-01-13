// Package envwrapper is a godotenv wrapper that retrieves .env variables from a
// map[slice]any and returns a new map with all the values in the correct type.
package envwrapper

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	errEnvMissing     = errors.New("required env variable missing or empty")
	errEnvInvalidInt  = errors.New("env variable must be an integer")
	errEnvInvalidBool = errors.New("env variable must be a boolean")
	errEnvUnsupported = errors.New("unsupported env variable type")
)

// Parse reads the structure of cfg, a map[string]any, and returns a pointer to a
// new map where keys = 'string' and values = the corresponding environment variable
// cast into the type stipulated by 'any'. The values of 'any' are the default fallbacks
// if Load cannot find or cast the environment variable.
//
// string, int, bool, []byte are supported.
//
// Example 1:
//
//	cfg := map[string]any{
//	    "MY_SERVER":   "localhost",
//	    "MY_PORT":     456,
//	    "MY_PASSWORD": []byte("default_value"),
//	    "USE_SSL":     true,
//	}
//	env := envwrapper.Parse(cfg)
//	(e.g.) sendMail(env["MY_SERVER"], env["MY_PORT"], env["MY_PASSWORD"])
//
// Also, required values without a stipulated default can be passed a nil pointer to that specific type:
// Example 2:
//
//	cfg := map[string]any{
//	    "SOME_REQUIRED_ENV_VAR":    (*string)(nil),
//	    "REQUIRED_KEY":             (*[]byte)(nil),
//	    "REQUIRED_BOOL_NO_DEFAULT": (*bool)(nil),
//	}
//	env := envwrapper.Parse(cfg)
//
// Any combination of example 1 and 2 is possible, where some values have a default fallback and some do not.
// Failure to load or cast any required variable will panic.
//
// Secrets, passwords, API keys, etc, should be saved as []bytes rather than strings,
// since strings are immutable in Go. Optionally, call defer envwrapper.WipeSecrets(env) in main to
// zero-out secrets on program close to prevent memory-dump leaks.
//
// Example 3:
//
//	...
//	env := envwrapper.Parse(cfg)
//	defer envwrapper.WipeSecrets(env)
//
// envwrapper loads any .env file in the current path. Filenames are optional to manually load one or more .env files.
func Parse(cfg map[string]any, filenames ...string) map[string]any {
	err := godotenv.Load(filenames...)
	if err != nil {
		panic(fmt.Errorf(".env not found: %w", err))
	}
	env := make(map[string]any, len(cfg))

	for key, defaultKey := range cfg {
		env[key] = resolve(key, defaultKey)
	}
	return env
}

// WipeSecrets should be called as a deferred function in main after Parse if the secrets, passwords, or
// API keys are passed into the env map. WipeSecrets zeros-out all instances of []byte and *[]byte and
// then deletes the corresponding map entry. It is used to safeguard again memory-dump leaks.
//
// Example:
//
//	...
//	env := envwrapper.Parse(cfg)
//	defer envwrapper.WipeSecrets(env)
func WipeSecrets(env *map[string]any) {
	if env == nil {
		return
	}
	for key, val := range *env {
		switch v := val.(type) {
		case []byte:
			for i := range v {
				v[i] = 0
			}
			delete(*env, key)
		case *[]byte:
			if v != nil {
				for i := range *v {
					(*v)[i] = 0
				}
			}
			delete(*env, key)
		}
	}
}

func resolve(key string, defaultKey any) any {
	rawString, keyExists := os.LookupEnv(key)
	rawString = strings.TrimSpace(rawString)

	switch defaultKeyType := defaultKey.(type) {
	case string:
		return resolveString(keyExists, rawString, defaultKeyType)
	case *string:
		return requireString(keyExists, rawString, key)
	case int:
		return resolveInt(keyExists, rawString, defaultKeyType)
	case *int:
		return requireInt(keyExists, rawString, key)
	case bool:
		return resolveBool(keyExists, rawString, defaultKeyType)
	case *bool:
		return requireBool(keyExists, rawString, key)
	case []byte:
		return resolveBytes(keyExists, rawString, defaultKeyType)
	case *[]byte:
		return requireBytes(keyExists, rawString, key)
	default:
		panic(fmt.Errorf("%w: %s (%T)", errEnvUnsupported, key, defaultKey))
	}
}

func resolveString(keyExists bool, rawString, defaultKey string) string {
	if !keyExists || rawString == "" {
		return defaultKey
	}
	return rawString
}

func requireString(keyExists bool, rawString, key string) string {
	if !keyExists || rawString == "" {
		panic(fmt.Errorf("%w: %s", errEnvMissing, key))
	}
	return rawString
}

func resolveInt(keyExists bool, rawString string, defaultKey int) int {
	if !keyExists || rawString == "" {
		return defaultKey
	}
	i, err := strconv.Atoi(rawString)
	if err != nil {
		return defaultKey
	}
	return i
}

func requireInt(keyExists bool, rawString, key string) int {
	if !keyExists || rawString == "" {
		panic(fmt.Errorf("%w: %s", errEnvMissing, key))
	}
	i, err := strconv.Atoi(rawString)
	if err != nil {
		panic(fmt.Errorf("%w: %s", errEnvInvalidInt, key))
	}
	return i
}

func resolveBool(keyExists bool, rawString string, defaultKey bool) bool {
	if !keyExists || rawString == "" {
		return defaultKey
	}
	boolean, err := strconv.ParseBool(rawString)
	if err != nil {
		return defaultKey
	}
	return boolean
}

func requireBool(keyExists bool, rawString, key string) bool {
	if !keyExists || rawString == "" {
		panic(fmt.Errorf("%w: %s", errEnvMissing, key))
	}
	boolean, err := strconv.ParseBool(rawString)
	if err != nil {
		panic(fmt.Errorf("%w: %s", errEnvInvalidBool, key))
	}
	return boolean
}

func resolveBytes(keyExists bool, rawString string, defaultKey []byte) []byte {
	if !keyExists || rawString == "" {
		return defaultKey
	}
	return []byte(rawString)
}

func requireBytes(keyExists bool, rawString, key string) []byte {
	if !keyExists || rawString == "" {
		panic(fmt.Errorf("%w: %s", errEnvMissing, key))
	}
	return []byte(rawString)
}
