package cmd

import (
	"log"
)

// Execute executes the commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
