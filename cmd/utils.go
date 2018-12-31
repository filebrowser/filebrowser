package cmd

import (
	"crypto/rand"
	"errors"
	"os"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/bolt"
	"github.com/filebrowser/filebrowser/types"
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

func getStore(db *storm.DB) *types.Store {
	usersStore := &types.UsersVerify{
		Store: bolt.UsersStore{
			DB: db,
		},
	}

	configSore := &types.ConfigVerify{
		Store: bolt.ConfigStore{
			DB:    db,
			Users: usersStore,
		},
	}

	return &types.Store{
		Users:  usersStore,
		Config: configSore,
		Share:  bolt.ShareStore{DB: db},
	}
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	checkErr(err)
	return b
}
