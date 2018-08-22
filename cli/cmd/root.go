package cmd

import (
	"log"
	"strings"

	fb "github.com/filebrowser/filebrowser/lib"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "filebrowser",
	Version: fb.Version,
	Aliases: []string{"serve"},
	Short:   "A stylish web-based file manager",
	Long: `Command 'serve' is the default. Filebrowser is started
with the provided envvars, flags and/or config file. For example:

filebrowser -c config.json -p 80 -s ./srv

File Browser is a static binary composed of a golang backend and
a Vue.js frontend to create, edit, copy, move, download your files
easily, everywhere, every time.`,
	//	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	checkRootAlias()
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetVersionTemplate("File Browser {{printf \"version %s\" .Version}}\n")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (defaults are './.filebrowser[ext]', '$HOME/.filebrowser[ext]' or '/etc/filebrowser/.filebrowser[ext]')")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		v.AddConfigPath(".")
		v.AddConfigPath(home)
		v.AddConfigPath("/etc/filebrowser/")
		v.SetConfigName(".filebrowser")
	} else {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	}

	v.SetEnvPrefix("FB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(v.ConfigParseError); ok {
			panic(err)
		}
	} else {
		log.Println("Using config file:", v.ConfigFileUsed())
	}
}
