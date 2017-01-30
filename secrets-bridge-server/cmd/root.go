package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "secrets-bridge-server",
	Short: "Serves secrets to Docker build runs, securely, over from the host",
	Long:  ``,
	// Uncomment for top-level action..
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var bridgeConfFilename string

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&bridgeConfFilename, "bridge-conf", "w", "bridge-conf", "Bridge authentication file. Written to in serve command, read in kill command. Defaults to `bridge-conf`")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".secrets-bridge-server")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("SECRETS_BRIDGE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
