package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/client"
	"github.com/spf13/cobra"
)

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kills the remote bridge server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		confFile := bridgeConfFilenameWithDefault()
		bridgeConf, err := ioutil.ReadFile(confFile)
		if err != nil {
			log.Fatalln("reading %q: %s", confFile, err)
		}

		bridge, err := bridge.NewFromString(string(bridgeConf))
		if err != nil {
			log.Fatalln("--bridge-conf has an invalid value:", err)
		}

		c := client.NewClient(bridge)
		if err := c.ChooseEndpoint(); err != nil {
			log.Fatalln("error pinging server:", err)
		}

		err = c.Quit()
		if err != nil {
			log.Fatalln("error killing previous server:", err)
		}

		fmt.Println("bridge server terminated")
	},
}

func init() {
	RootCmd.AddCommand(killCmd)
}
