package cmd

import (
	"crypto/rand"
	"errors"
	"os"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/bolt"
	"github.com/spf13/cobra"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func mustGetString(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	checkErr(err)
	return s
}

func mustGetBool(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	checkErr(err)
	return b
}

func getDB() *storm.DB {
	if _, err := os.Stat(databasePath); err != nil {
		panic(errors.New(databasePath + " does not exist. Please run 'filebrowser init' first."))
	}

	db, err := bolt.Open(databasePath)
	checkErr(err)
	return db
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	checkErr(err)
	return b
}
