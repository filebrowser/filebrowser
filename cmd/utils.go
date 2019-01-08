package cmd

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
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
		fmt.Println(err)
		os.Exit(1)
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

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	checkErr(err)
	// Note that err == nil only if we read len(b) bytes.
	return b
}

type cobraFunc func(cmd *cobra.Command, args []string)
type pythonFunc func(cmd *cobra.Command, args []string, data pythonData)

type pythonConfig struct {
	noDB      bool
	allowNoDB bool
}

type pythonData struct {
	hadDB bool
	store *storage.Storage
}

func python(fn pythonFunc, cfg pythonConfig) cobraFunc {
	return func(cmd *cobra.Command, args []string) {
		data := pythonData{hadDB: true}

		path := v.GetString("database")
		_, err := os.Stat(path)

		if os.IsNotExist(err) {
			data.hadDB = false

			if !cfg.noDB && !cfg.allowNoDB {
				log.Fatal(path + " does not exist. Please run 'filebrowser config init' first.")
			}
		} else if err != nil {
			panic(err)
		} else if err == nil && cfg.noDB {
			log.Fatal(path + " already exists")
		}

		db, err := storm.Open(path)
		checkErr(err)
		defer db.Close()
		data.store = bolt.NewStorage(db)
		fn(cmd, args, data)
	}
}

func marshal(filename string, data interface{}) error {
	fd, err := os.Create(filename)
	checkErr(err)
	defer fd.Close()

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "    ")
		return encoder.Encode(data)
	case ".yml", ".yaml":
		encoder := yaml.NewEncoder(fd)
		return encoder.Encode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

func unmarshal(filename string, data interface{}) error {
	fd, err := os.Open(filename)
	checkErr(err)
	defer fd.Close()

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		return json.NewDecoder(fd).Decode(data)
	case ".yml", ".yaml":
		return yaml.NewDecoder(fd).Decode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

func jsonYamlArg(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	switch ext := filepath.Ext(args[0]); ext {
	case ".json", ".yml", ".yaml":
		return nil
	default:
		return errors.New("invalid format: " + ext)
	}
}

func cleanUpInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpInterfaceArray(in []interface{}) []interface{} {
	result := make([]interface{}, len(in))
	for i, v := range in {
		result[i] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanUpInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanUpInterfaceMap(v)
	default:
		return v
	}
}
