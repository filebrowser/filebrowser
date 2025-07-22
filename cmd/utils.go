package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
)

const dbPerms = 0640

func returnErr(err error) error {
	if err != nil {
		return err
	}
	return nil
}

func getString(flags *pflag.FlagSet, flag string) (string, error) {
	s, err := flags.GetString(flag)
	return s, returnErr(err)
}

func getMode(flags *pflag.FlagSet, flag string) (fs.FileMode, error) {
	s, err := getString(flags, flag)
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, err
	}
	return fs.FileMode(b), nil
}

func getBool(flags *pflag.FlagSet, flag string) (bool, error) {
	b, err := flags.GetBool(flag)
	return b, returnErr(err)
}

func getUint(flags *pflag.FlagSet, flag string) (uint, error) {
	b, err := flags.GetUint(flag)
	return b, returnErr(err)
}

func generateKey() []byte {
	k, err := settings.GenerateKey()
	if err != nil {
		panic(err)
	}
	return k
}

type cobraFunc func(cmd *cobra.Command, args []string) error
type pythonFunc func(cmd *cobra.Command, args []string, data *pythonData) error

type pythonConfig struct {
	noDB      bool
	allowNoDB bool
}

type pythonData struct {
	hadDB bool
	store *storage.Storage
	err   error
}

func dbExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.Size() != 0, nil
	}

	if os.IsNotExist(err) {
		d := filepath.Dir(path)
		_, err = os.Stat(d)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0700); err != nil { //nolint:govet
				return false, err
			}
			return false, nil
		}
	}

	return false, err
}

func python(fn pythonFunc, cfg pythonConfig) cobraFunc {
	return func(cmd *cobra.Command, args []string) error {
		data := &pythonData{hadDB: true}

		path := getStringParam(cmd.Flags(), "database")
		absPath, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		exists, err := dbExists(path)

		if err != nil {
			panic(err)
		} else if exists && cfg.noDB {
			log.Fatal(absPath + " already exists")
		} else if !exists && !cfg.noDB && !cfg.allowNoDB {
			log.Fatal(absPath + " does not exist. Please run 'filebrowser config init' first.")
		} else if !exists && !cfg.noDB {
			log.Println("Warning: filebrowser.db can't be found. Initialing in " + strings.TrimSuffix(absPath, "filebrowser.db"))
		}

		log.Println("Using database: " + absPath)
		data.hadDB = exists
		db, err := storm.Open(path, storm.BoltOptions(dbPerms, nil))
		if err != nil {
			return err
		}
		defer db.Close()
		data.store, err = bolt.NewStorage(db)
		if err != nil {
			return err
		}
		return fn(cmd, args, data)
	}
}

func marshal(filename string, data interface{}) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
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
	if err != nil {
		return err
	}
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

// convertCmdStrToCmdArray checks if cmd string is blank (whitespace included)
// then returns empty string array, else returns the split word array of cmd.
// This is to ensure the result will never be []string{""}
func convertCmdStrToCmdArray(cmd string) []string {
	var cmdArray []string
	trimmedCmdStr := strings.TrimSpace(cmd)
	if trimmedCmdStr != "" {
		cmdArray = strings.Split(trimmedCmdStr, " ")
	}
	return cmdArray
}
