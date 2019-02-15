package cmd

import (
	"fmt"

	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hashCmd)
}

var hashCmd = &cobra.Command{
	Use:   "hash <password>",
	Short: "Hashes a password",
	Long:  `Hashes a password using bcrypt algorithm.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := users.HashPwd(args[0])
		checkErr(err)
		fmt.Println(pwd)
	},
}
