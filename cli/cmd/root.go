package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "filebrowser",
	Version: "(untracked)",
	Short: "A stylish web-based file manager",
	Long: `File Browser is static binary composed of a golang backend
and a Vue.js frontend to create, edit, copy, move, download your files
easily, everywhere, every time.`,
/*
	Run: func(cmd *cobra.Command, args []string) {
		serveCmd.Run(cmd, args)
	},
*/
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetVersionTemplate("File Browser {{printf \"version %s\" .Version}}\n")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/filebrowser/")
		viper.SetConfigName(".filebrowser")
	}

	viper.SetEnvPrefix("FB")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			panic(err)
		}
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

/*
	// Add a configuration file if set.
	if config != "" {
		ext := filepath.Ext(config)
		dir := filepath.Dir(config)
		config = strings.TrimSuffix(config, ext)

		if dir != "" {
			viper.AddConfigPath(dir)
			config = strings.TrimPrefix(config, dir)
		}

		viper.SetConfigName(config)
	}
*/