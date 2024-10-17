package cmd

import (
	"embed"
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tink3rlabs/magic/logger"
)

var ConfigFS embed.FS
var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "ToDo is a reference implementaion of a common service architecture",
	Long: `ToDo is a reference implementaion of a common service architecture brought to you with love by tink3rlabs.
Complete documentation is available at https://github.com/tink3rlabs/todo-service`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.todo.yaml)")
	if viperBindFlagsErr := viper.BindPFlags(rootCmd.Flags()); viperBindFlagsErr != nil {
		fmt.Println(viperBindFlagsErr)
		os.Exit(1)
	}
	rootCmd.AddCommand(serverCommand)
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
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

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".todo")
	}

	viper.SetEnvPrefix("TODO")
	viper.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	config := loggerConfig()
	logger.Init(config)
}

func loggerConfig() *logger.Config {
	// Fetch the log level and format from the config file
	levelStr := viper.GetString("logger.level")
	json := viper.GetBool("logger.json")

	return &logger.Config{
		Level: logger.MapLogLevel(levelStr),
		JSON:  json,
	}
}
