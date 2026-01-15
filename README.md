# envwrapper

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/glenntam/envwrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/glenntam/envwrapper)](https://goreportcard.com/report/github.com/glenntam/envwrapper)

Package envwrapper is a godotenv wrapper that retrieves .env variables from a
map[slice]any and returns a new map with all the values in the correct type.

## Functions

### func [Parse](https://github.com/glenntam/envwrapper/blob/main/envwrapper.go#L64)

`func Parse(cfg map[string]any, filenames ...string) map[string]any`

Parse reads the structure of cfg, a map[string]any, and returns a pointer to a
new map where keys = 'string' and values = the corresponding environment variable
cast into the type stipulated by 'any'. The values of 'any' are the default fallbacks
if Load cannot find or cast the environment variable.

string, int, bool, []byte are supported.

Example 1:

```go
cfg := map[string]any{
    "MY_SERVER":   "localhost",
    "MY_PORT":     456,
    "MY_PASSWORD": []byte("default_value"),
    "USE_SSL":     true,
}
env := envwrapper.Parse(cfg)
(e.g.) sendMail(env["MY_SERVER"].(string), env["MY_PORT"].(int), env["MY_PASSWORD"].([]byte), env["USE_SSL"].(bool))
```

Also, required values without a stipulated default can be passed a nil pointer to that specific type:
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
Failure to load or cast any required variable will panic.

Secrets, passwords, API keys, etc, should be saved as []bytes rather than strings,
since strings are immutable in Go. Optionally, call defer envwrapper.WipeSecrets(env) in main to
zero-out secrets on program close to prevent memory-dump leaks.

Example 3:

```go
...
env := envwrapper.Parse(cfg)
defer envwrapper.WipeSecrets(env)
```

envwrapper loads any .env file in the current path. Filenames are optional to manually load one or more .env files.

### func [WipeSecrets](https://github.com/glenntam/envwrapper/blob/main/envwrapper.go#L86)

`func WipeSecrets(env *map[string]any)`

WipeSecrets should be called as a deferred function in main after Parse if the secrets, passwords, or
API keys are passed into the env map. WipeSecrets zeros-out all instances of []byte and *[]byte and
then deletes the corresponding map entry. It is used to safeguard again memory-dump leaks.

Example:

```go
...
env := envwrapper.Parse(cfg)
defer envwrapper.WipeSecrets(env)
```
