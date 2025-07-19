package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	rootCmd.AddCommand(hashCmd)
}

var hashCmd = &cobra.Command{
	Use:   "hash <password>",
	Short: "Hashes a password",
	Long:  `Hashes a password using bcrypt algorithm.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		pwd, err := users.HashPwd(args[0])
		if err != nil {
			return err
		}
		fmt.Println(pwd)
		return nil
	},
}
