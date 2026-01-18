# envwrapper

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/glenntam/envwrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/glenntam/envwrapper)](https://goreportcard.com/report/github.com/glenntam/envwrapper)

Package envwrapper is a godotenv convenience wrapper. It's designed so user code can define fallbacks
and assign each variable's data type with a single map[string]any at compile time with zero surprises.


## Types

### type [EnvParsed](https://github.com/glenntam/envwrapper/blob/main/envwrapper.go#L26)

```go
type EnvParsed struct {
	Str   map[string]string
	Int   map[string]int
	Bool  map[string]bool
	Bytes map[string][]byte
}
```

EnvParsed returns all the environment variable values categorized by type.
The user-supplied cfg default values are used if the environment values are invalid.

## Functions

#### func [Parse](https://github.com/glenntam/envwrapper/blob/main/envwrapper.go#L77)

`func Parse(cfg map[string]any, filenames ...string) *EnvParsed`

Parse reads the structure of cfg, a map[string]any, and returns a pointer to a
struct (*EnvParsed) with environment variables categorized by type.

string, int, bool, []byte are supported.

Values in cfg are app-defined defaults, in case the environment variable doesn't exist or is invalid.

Example 1:

```go
cfg := map[string]any{
    "MY_SERVER":   "localhost",
    "MY_PORT":     456,
    "MY_PASSWORD": []byte("default_value"),
    "USE_SSL":     true,
}
env := envwrapper.Parse(cfg)
(e.g.) sendMail(env.Str["MY_SERVER"], env.Int["MY_PORT"], env.Bytes["MY_PASSWORD"], env.Bool["USE_SSL"])
```

Also, required values without a default value can be passed a nil pointer to that specific type.

Example 2:

```go
cfg := map[string]any{
    "SOME_REQUIRED_ENV_VAR":    (*string)(nil),
    "REQUIRED_KEY":             (*[]byte)(nil),
    "REQUIRED_BOOL_NO_DEFAULT": (*bool)(nil),
}
env := envwrapper.Parse(cfg)
```

Any combination of example 1 and 2 is possible, where some values have a default fallback and some do not.
By design, failure to load or cast any required variable will panic.

Secrets, passwords, API keys, etc, should be saved as []bytes rather than strings since
strings are immutable and value-copied in Go. Optionally, call defer envwrapper.WipeSecrets(env)
to zero-out secrets on program close to prevent memory-dump leaks.

Example 3:

```go
...
env := envwrapper.Parse(cfg)
defer envwrapper.WipeSecrets(env)
```

envwrapper loads any .env file in the current path. Filenames are optional to manually load one or more .env files.
Duplicated environment variables will be overriden in the order the files are loaded, like godotenv.

### func [WipeSecrets](https://github.com/glenntam/envwrapper/blob/main/envwrapper.go#L105)

`func WipeSecrets(env *EnvParsed)`

WipeSecrets should be called as a deferred function in main after Parse() if any
secrets, passwords, or API keys are used. WipeSecrets zeros-out and deletes all
instances of []byte in EnvParsed. It is used to safeguard again memory-dump leaks.

Example:

```go
...
env := envwrapper.Parse(cfg)
defer envwrapper.WipeSecrets(env)
```
