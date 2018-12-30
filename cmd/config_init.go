package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/filebrowser/filebrowser/bolt"
	"github.com/filebrowser/filebrowser/types"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configInitCmd)
	addConfigFlags(configInitCmd)
	configInitCmd.MarkFlagRequired("scope")
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new database",
	Long: `Initialize a new database to use with File Browser. All of
this options can be changed in the future with the command
"filebrowser config set". The user related flags apply
to the defaults when creating new users and you don't
override the options.`,
	Args: cobra.NoArgs,
	Run:  initDatabase,
}

func initDatabase(cmd *cobra.Command, args []string) {
	if _, err := os.Stat(databasePath); err == nil {
		panic(errors.New(databasePath + " already exists"))
	}

	defaults := types.UserDefaults{}
	getUserDefaults(cmd, &defaults, true)
	authMethod, auther := getAuthentication(cmd)

	settings := &types.Settings{
		Key:        generateRandomBytes(64), // 256 bits
		BaseURL:    mustGetString(cmd, "baseURL"),
		Signup:     mustGetBool(cmd, "signup"),
		Defaults:   defaults,
		AuthMethod: authMethod,
	}

	runner := &types.Runner{
		Commands: map[string][]string{},
	}

	for _, event := range types.DefaultEvents {
		runner.Commands[event] = []string{}
	}

	db, err := bolt.Open(databasePath)
	checkErr(err)
	defer db.Close()

	st := bolt.ConfigStore{DB: db}
	err = st.SaveSettings(settings)
	checkErr(err)
	err = st.SaveRunner(runner)
	checkErr(err)
	err = st.SaveAuther(auther)
	checkErr(err)

	fmt.Printf(`
Congratulations! You've set up your database to use with File Browser.
Now add your first user via 'filebrowser users new' and then you just
need to call the main command to boot up the server.
`)
	printSettings(settings, auther)
}
