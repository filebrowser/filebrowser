package cmd

import (
	"crypto/rand"
	"errors"
	"log"
	"os"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
)

func vaddP(f *pflag.FlagSet, k, p string, i interface{}, u string) {
	switch y := i.(type) {
	case bool:
		f.BoolP(k, p, y, u)
	case int:
		f.IntP(k, p, y, u)
	case string:
		f.StringP(k, p, y, u)
	}
	v.SetDefault(k, i)
}

func vadd(f *pflag.FlagSet, k string, i interface{}, u string) {
	switch y := i.(type) {
	case bool:
		f.Bool(k, y, u)
	case int:
		f.Int(k, y, u)
	case string:
		f.String(k, y, u)
	}
	v.SetDefault(k, i)
}

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

func mustGetUint(cmd *cobra.Command, flag string) uint {
	b, err := cmd.Flags().GetUint(flag)
	checkErr(err)
	return b
}

func getDB() *storm.DB {
	databasePath := v.GetString("database")
	if _, err := os.Stat(databasePath); err != nil {
		panic(errors.New(databasePath + " does not exist. Please run 'filebrowser init' first."))
	}

	db, err := storm.Open(databasePath)
	checkErr(err)
	return db
}

func getStorage(db *storm.DB) *storage.Storage {
	return bolt.NewStorage(db)
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	checkErr(err)
	// Note that err == nil only if we read len(b) bytes.
	return b
}

type cobraFunc func(cmd *cobra.Command, args []string)
type pythonFunc func(cmd *cobra.Command, args []string, st *storage.Storage)

type pythonConfig struct {
	noDB bool
}

func python(fn pythonFunc, cfg pythonConfig) cobraFunc {
	return func(cmd *cobra.Command, args []string) {
		path := v.GetString("database")
		_, err := os.Stat(path)

		if err != nil && !os.IsNotExist(err) {
			panic(err)
		} else if err != nil && !cfg.noDB {
			log.Fatal(path + " does not exist. Please run 'filebrowser config init' first.")
		} else if err == nil && cfg.noDB {
			log.Fatal(path + " already exists")
		}

		db, err := storm.Open(path)
		checkErr(err)
		defer db.Close()
		sto := bolt.NewStorage(db)
		fn(cmd, args, sto)
	}
}
