package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
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

// newHookAuth builds a HookAuth for a freshly provisioned user with the given
// parsed hook fields (hook.action=auth is added automatically).
func newHookAuth(store *mockUserStore, s *settings.Settings, srv *settings.Server, username string, fields map[string]string) *HookAuth {
	fields["hook.action"] = "auth"
	return &HookAuth{
		Users:    store,
		Settings: s,
		Server:   srv,
		Cred:     hookCred{Username: username, Password: "a-strong-password"},
		Fields:   hookFields{Values: fields},
	}
}

// With CreateUserDir enabled and no explicit scope from the hook, a provisioned
// hook user must receive its own home directory rather than the server root.
func TestHookSaveUserCreateUserDirIsolatesScope(t *testing.T) {
	t.Parallel()

	store := &mockUserStore{users: make(map[string]*users.User)}
	srv := &settings.Server{Root: t.TempDir()}
	s := &settings.Settings{
		Key:              []byte("key"),
		CreateUserDir:    true,
		UserHomeBasePath: "/users",
		Defaults: settings.UserDefaults{
			Scope: ".",
			Perm:  users.Permissions{Create: true},
		},
	}

	u, err := newHookAuth(store, s, srv, "alice", map[string]string{}).SaveUser()
	if err != nil {
		t.Fatalf("SaveUser error: %v", err)
	}
	if u.Scope != "/users/alice" {
		t.Errorf("hook user without explicit scope: expected /users/alice, got %q", u.Scope)
	}
}

// A scope explicitly returned by the hook takes precedence over the automatic
// per-user home directory derivation.
func TestHookSaveUserRespectsExplicitScope(t *testing.T) {
	t.Parallel()

	store := &mockUserStore{users: make(map[string]*users.User)}
	srv := &settings.Server{Root: t.TempDir()}
	s := &settings.Settings{
		Key:              []byte("key"),
		CreateUserDir:    true,
		UserHomeBasePath: "/users",
		Defaults: settings.UserDefaults{
			Scope: ".",
			Perm:  users.Permissions{Create: true},
		},
	}

	u, err := newHookAuth(store, s, srv, "teamlead", map[string]string{"user.scope": "/shared/team"}).SaveUser()
	if err != nil {
		t.Fatalf("SaveUser error: %v", err)
	}
	if u.Scope != "/shared/team" {
		t.Errorf("explicit hook scope should win, got %q", u.Scope)
	}
}
