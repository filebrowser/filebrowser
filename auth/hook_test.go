package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeHookScript writes a POSIX shell script to a temp file and returns its
// path, marking it executable.
func writeHookScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hook.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o700); err != nil {
		t.Fatalf("failed to write hook script: %v", err)
	}
	return path
}

// TestRunCommandNoCredentialInjection ensures that attacker-controlled
// credentials submitted at the unauthenticated login endpoint cannot be
// injected into the hook command string. Credentials must only ever reach the
// hook through the USERNAME/PASSWORD environment variables, never via string
// substitution into the command itself (CWE-78/CWE-88).
func TestRunCommandNoCredentialInjection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX shell")
	}

	marker := filepath.Join(t.TempDir(), "pwned")

	// The hook simply blocks. If the credential were ever interpolated into the
	// command string and evaluated by a shell, the embedded `touch` would
	// create the marker file.
	script := writeHookScript(t, "echo hook.action=block\n")

	a := &HookAuth{
		Command: script,
		Cred: hookCred{
			Username: `"; touch ` + marker + `; #`,
			Password: `$(touch ` + marker + `)`,
		},
	}

	action, err := a.RunCommand()
	if err != nil {
		t.Fatalf("RunCommand returned error: %v", err)
	}
	if action != "block" {
		t.Fatalf("expected action %q, got %q", "block", action)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("credential injection executed: marker file %q was created", marker)
	}
}

// TestRunCommandReceivesCredentialsViaEnv verifies the supported contract: the
// hook receives credentials through the USERNAME and PASSWORD environment
// variables.
func TestRunCommandReceivesCredentialsViaEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX shell")
	}

	script := writeHookScript(t, `if [ "$USERNAME" = alice ] && [ "$PASSWORD" = secret ]; then
  echo hook.action=auth
else
  echo hook.action=block
fi
`)

	a := &HookAuth{
		Command: script,
		Cred: hookCred{
			Username: "alice",
			Password: "secret",
		},
	}

	action, err := a.RunCommand()
	if err != nil {
		t.Fatalf("RunCommand returned error: %v", err)
	}
	if action != "auth" {
		t.Fatalf("expected action %q, got %q", "auth", action)
	}
}
