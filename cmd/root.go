// Copyright © 2017 Alexandre Bourget <alex@bourget.cc>

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "secrets-bridge",
	Short: "Manage secrets securely across a network boundary.",
	Long:  `See https://github.com/abourget/secrets-bridge for details.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var bridgeConf string
var bridgeConfFilename string

func bridgeConfFilenameWithDefault() string {
	if bridgeConfFilename != "" {
		return bridgeConfFilename
	}
	return filepath.Join(os.Getenv("HOME"), ".bridge-conf")
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".secrets-bridge")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("SECRETS_BRIDGE")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func newClient(bridgeConf string) (*client.Client, error) {
	var brConf *bridge.Bridge

	if bridgeConf == "" {
		br, err := bridge.NewFromDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("couldn't load bridge conf: %s", err)
		}
		brConf = br
	} else {
		br, err := bridge.NewFromString(bridgeConf)
		if err != nil {
			return nil, fmt.Errorf("--bridge-conf has an invalid value: %s", err)
		}

		brConf = br
	}

	c := client.NewClient(brConf)

	err := c.ChooseEndpoint()
	if err != nil {
		return c, fmt.Errorf("error pinging server: %s", err)
	}

	return c, nil
}
