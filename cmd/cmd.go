package cmd

import (
	"fmt"
	"os"
)

// Execute executes the commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
