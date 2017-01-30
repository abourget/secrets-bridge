// Copyright © 2017 Alexandre Bourget <alex@bourget.cc>

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "secrets-bridge-client",
	Short: "Connect to a secrets server and expose them locally",
	Long: `The secrets-bridge-client can also forward SSH-Agents from over the bridge.

This process is to be run from inside a Docker container, and be fed a
--bridge-conf string that can be passed through --build-args to the
'docker build' run.

See https://github.com/abourget/secrets-bridge for details.`,
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP("bridge-conf", "w", "", "Base64-encoded Bridge `configuration`.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".secrets-bridge-client")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("SECRETS_BRIDGE")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
