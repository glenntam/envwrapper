package envwrapper

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// --- helpers ---

func withDotEnv(t *testing.T, contents string) {
	t.Helper()

	tmp := t.TempDir()
	t.Chdir(tmp)

	// Completely unset all existing env vars
	for _, e := range os.Environ() {
		if k, _, ok := strings.Cut(e, "="); ok {
			_ = os.Unsetenv(k)
		}
	}

	if err := os.WriteFile(".env", []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustPanic(t *testing.T, fn func()) error {
	t.Helper()

	var got error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					got = err
				} else {
					t.Fatalf("panic was not error: %T", r)
				}
			}
		}()
		fn()
	}()

	if got == nil {
		t.Fatalf("expected panic, got none")
	}
	return got
}

// --- tests ---

func TestParse_OptionalValues(t *testing.T) {
	withDotEnv(t, `
STR=value
INT=123
BOOL=false
`)

	cfg := map[string]any{
		"STR":  "default",
		"INT":  0,
		"BOOL": true,
	}

	env := Parse(cfg)

	if (*env)["STR"] != "value" {
		t.Fatalf("string mismatch")
	}
	if (*env)["INT"] != 123 {
		t.Fatalf("int mismatch")
	}
	if (*env)["BOOL"] != false {
		t.Fatalf("bool mismatch")
	}
}

func TestParse_OptionalFallbacks(t *testing.T) {
	withDotEnv(t, `
STR=
INT=not-an-int
BOOL=not-a-bool
`)

	cfg := map[string]any{
		"STR":  "default",
		"INT":  42,
		"BOOL": true,
	}

	env := Parse(cfg)

	if (*env)["STR"] != "default" {
		t.Fatalf("expected default string")
	}
	if (*env)["INT"] != 42 {
		t.Fatalf("expected default int")
	}
	if (*env)["BOOL"] != true {
		t.Fatalf("expected default bool")
	}
}

func TestParse_OptionalBytes(t *testing.T) {
	withDotEnv(t, `
BYTES=secret
`)

	cfg := map[string]any{
		"BYTES": []byte("default"),
	}

	env := Parse(cfg)

	val, ok := (*env)["BYTES"]
	if !ok {
		t.Fatalf("BYTES missing")
	}

	b, ok := val.([]byte)
	if !ok {
		t.Fatalf("wrong type: %T", val)
	}
	if string(b) != "secret" {
		t.Fatalf("unexpected value")
	}
}

func TestParse_RequiredMissingPanics(t *testing.T) {
	withDotEnv(t, ``)

	cfg := map[string]any{
		"REQ": (*string)(nil),
	}

	err := mustPanic(t, func() {
		Parse(cfg)
	})

	if !errors.Is(err, errEnvMissing) {
		t.Fatalf("expected errEnvMissing, got %v", err)
	}
}

func TestParse_RequiredInvalidIntPanics(t *testing.T) {
	withDotEnv(t, `
REQ=abc
`)

	cfg := map[string]any{
		"REQ": (*int)(nil),
	}

	err := mustPanic(t, func() {
		Parse(cfg)
	})

	if !errors.Is(err, errEnvInvalidInt) {
		t.Fatalf("expected errEnvInvalidInt, got %v", err)
	}
}

func TestParse_RequiredInvalidBoolPanics(t *testing.T) {
	withDotEnv(t, `
REQ=abc
`)

	cfg := map[string]any{
		"REQ": (*bool)(nil),
	}

	err := mustPanic(t, func() {
		Parse(cfg)
	})

	if !errors.Is(err, errEnvInvalidBool) {
		t.Fatalf("expected errEnvInvalidBool, got %v", err)
	}
}

func TestParse_RequiredBytes(t *testing.T) {
	withDotEnv(t, `
BYTES=secret
`)

	cfg := map[string]any{
		"BYTES": (*[]byte)(nil),
	}

	env := Parse(cfg)

	val := (*env)["BYTES"]
	b, ok := val.([]byte)
	if !ok {
		t.Fatalf("wrong type: %T", val)
	}
	if string(b) != "secret" {
		t.Fatalf("unexpected bytes")
	}
}

func TestParse_UnsupportedTypePanics(t *testing.T) {
	withDotEnv(t, ``)

	cfg := map[string]any{
		"BAD": float64(1.23),
	}

	err := mustPanic(t, func() {
		Parse(cfg)
	})

	if !errors.Is(err, errEnvUnsupported) {
		t.Fatalf("expected errEnvUnsupported, got %v", err)
	}
}

func TestWipeSecrets(t *testing.T) {
	secret := []byte("topsecret")

	env := map[string]any{
		"A": secret,
		"B": 123,
	}

	WipeSecrets(&env)

	if _, ok := env["A"]; ok {
		t.Fatalf("secret key not deleted")
	}
	if env["B"] != 123 {
		t.Fatalf("non-secret key modified")
	}
	for _, b := range secret {
		if b != 0 {
			t.Fatalf("secret not zeroed")
		}
	}
}
