package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kills the remote bridge server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := newClient(bridgeConf)
		if err != nil {
			log.Fatalln(err)
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
	killCmd.Flags().StringVarP(&bridgeConf, "bridge-conf", "c", "", "Base64-encoded Bridge `configuration`.")
}
