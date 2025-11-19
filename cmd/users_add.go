package cmd

import (
	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	usersCmd.AddCommand(usersAddCmd)
	addUserFlags(usersAddCmd.Flags())
}

var usersAddCmd = &cobra.Command{
	Use:   "add <username> <password>",
	Short: "Create a new user",
	Long:  `Create a new user and add it to the database.`,
	Args:  cobra.ExactArgs(2),
	RunE: withStore(func(cmd *cobra.Command, args []string, st *store) error {
		flags := cmd.Flags()
		s, err := st.Settings.Get()
		if err != nil {
			return err
		}
		err = getUserDefaults(flags, &s.Defaults, false)
		if err != nil {
			return err
		}

		password, err := users.ValidateAndHashPwd(args[1], s.MinimumPasswordLength)
		if err != nil {
			return err
		}

		user := &users.User{
			Username: args[0],
			Password: password,
		}

		user.LockPassword, err = flags.GetBool("lockPassword")
		if err != nil {
			return err
		}

		user.DateFormat, err = flags.GetBool("dateFormat")
		if err != nil {
			return err
		}

		user.HideDotfiles, err = flags.GetBool("hideDotfiles")
		if err != nil {
			return err
		}

		s.Defaults.Apply(user)

		servSettings, err := st.Settings.GetServer()
		if err != nil {
			return err
		}
		// since getUserDefaults() polluted s.Defaults.Scope
		// which makes the Scope not the one saved in the db
		// we need the right s.Defaults.Scope here
		s2, err := st.Settings.Get()
		if err != nil {
			return err
		}

		userHome, err := s2.MakeUserDir(user.Username, user.Scope, servSettings.Root)
		if err != nil {
			return err
		}
		user.Scope = userHome

		err = st.Users.Save(user)
		if err != nil {
			return err
		}
		printUsers([]*users.User{user})
		return nil
	}, storeOptions{}),
}
