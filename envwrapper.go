// Package envwrapper is a godotenv wrapper that retrieves .env variables from a
// map[string]any and returns a struct categorized by data type.
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
	errEnvMissing        = errors.New("required env variable missing or empty")
	errEnvInvalidInt     = errors.New("env variable must be an integer")
	errEnvInvalidBool    = errors.New("env variable must be a boolean")
	errEnvUnsupported    = errors.New("unsupported env variable type")
	errEnvNilUnsupported = errors.New("cannot use 'nil' as a required type. " +
		"Only (*string)(nil), (*int)(nil), (*bool)(nil), (*[]byte)(nil) are supported")
)

// EnvParsed contains all the environment variables or the default values categorized by type.
type EnvParsed struct {
	Str   map[string]string
	Int   map[string]int
	Bool  map[string]bool
	Bytes map[string][]byte
}

// Parse reads the structure of cfg, a map[string]any, and returns a pointer to a
// struct of parsed environment variables, categorized by type.
//
// string, int, bool, []byte are supported.
//
// Values in cfg are app-defined defaults, in case the environment variable doesn't exist.
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
//	(e.g.) sendMail(env.Str["MY_SERVER"], env.Int["MY_PORT"], env.Bytes["MY_PASSWORD"], env.Bool["USE_SSL"])
//
// Also, required values without a default value can be passed a nil pointer to that specific type:
//
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
func Parse(cfg map[string]any, filenames ...string) *EnvParsed {
	err := godotenv.Load(filenames...)
	if err != nil {
		panic(fmt.Errorf(".env not found: %w", err))
	}

	env := &EnvParsed{
		Str:   make(map[string]string),
		Int:   make(map[string]int),
		Bool:  make(map[string]bool),
		Bytes: make(map[string][]byte),
	}

	for key, defaultKey := range cfg {
		resolve(key, defaultKey, env)
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
func WipeSecrets(env *EnvParsed) {
	if env == nil {
		return
	}
	for key, val := range env.Bytes {
		for i := range val {
			val[i] = 0
		}
		delete(env.Bytes, key)
	}
}

func resolve(key string, defaultKey any, env *EnvParsed) {
	if defaultKey == nil {
		panic(fmt.Errorf("%w", errEnvNilUnsupported))
	}
	rawString, keyExists := os.LookupEnv(key)
	rawString = strings.TrimSpace(rawString)

	switch defaultKeyType := defaultKey.(type) {
	case string:
		env.Str[key] = resolveString(keyExists, rawString, defaultKeyType)
	case *string:
		env.Str[key] = requireString(keyExists, rawString, key)
	case int:
		env.Int[key] = resolveInt(keyExists, rawString, defaultKeyType)
	case *int:
		env.Int[key] = requireInt(keyExists, rawString, key)
	case bool:
		env.Bool[key] = resolveBool(keyExists, rawString, defaultKeyType)
	case *bool:
		env.Bool[key] = requireBool(keyExists, rawString, key)
	case []byte:
		env.Bytes[key] = resolveBytes(keyExists, rawString, defaultKeyType)
	case *[]byte:
		env.Bytes[key] = requireBytes(keyExists, rawString, key)
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
