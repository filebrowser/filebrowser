package cmd

import (
	"testing"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
)

// TestEnvCollisions ensures that there are no collisions in the produced environment
// variable names for all commands and their flags.
func TestEnvCollisions(t *testing.T) {
	testEnvCollisions(t, rootCmd)
}

func testEnvCollisions(t *testing.T, cmd *cobra.Command) {
	for _, cmd := range cmd.Commands() {
		testEnvCollisions(t, cmd)
	}

	replacements := generateEnvKeyReplacements(cmd)
	envVariables := []string{}

	for i := range replacements {
		if i%2 != 0 {
			envVariables = append(envVariables, replacements[i])
		}
	}

	duplicates := lo.FindDuplicates(envVariables)

	if len(duplicates) > 0 {
		t.Errorf("Found duplicate environment variable keys for command %q: %v", cmd.Name(), duplicates)
	}
}

// TestGetSettingsFollowExternalSymlinks ensures that the followExternalSymlinks
// flag is persisted to the server config when set via "config set".
func TestGetSettingsFollowExternalSymlinks(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	addConfigFlags(flags)

	if err := flags.Parse([]string{"--followExternalSymlinks"}); err != nil {
		t.Fatal(err)
	}

	set := &settings.Settings{AuthMethod: auth.MethodJSONAuth}
	ser := &settings.Server{}

	if _, err := getSettings(flags, set, ser, &auth.JSONAuth{}, false); err != nil {
		t.Fatal(err)
	}

	if !ser.FollowExternalSymlinks {
		t.Error("expected FollowExternalSymlinks to be persisted as true")
	}
}
