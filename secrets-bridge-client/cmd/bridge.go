// Copyright © 2017 Alexandre Bourget <alex@bourget.cc>

package cmd

import (
	"fmt"
	"log"

	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/client"
	"github.com/spf13/cobra"
)

// bridgeCmd represents the bridge command
var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Setup bridge, connect to server and proxy the SSH-Agent or serve secrets securely.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		bridgeConf, err := cmd.Flags().GetString("bridge-conf")
		if err != nil {
			log.Fatalln("--bridge-conf invalid:", err)
		}

		bridge, err := bridge.NewFromString(bridgeConf)
		if err != nil {
			log.Fatalln("--bridge-conf has an invalid value:", err)
		}

		c := client.NewClient(bridge)
		err = c.Ping()
		if err != nil {
			log.Fatalln("error pinging server:", err)
		}

		fmt.Println("bridge server responding")

		// setup the SSH agent if we specified `--ssh-agent`, with websocket and all..
		// setup the listener BEFORE exec'ing the subprocess
		// exec.Command(args...)
		// quit when we're done, delete the unix socket is we were doing ssh-agent.

		// client --listen, forwards to the Client server, .. pass-through
		// handle secrets requests, which will be merely forwarded,
		// and b64 things will be asked right here.. we'll merely
		// forward from the client when he listen on `--listen`.. both
		// status codes and content.
		//

		// listen on 127.0.0.1:4444 or --listen port
		// hook the client.RequestProxier

		fmt.Println("bridge called")
	},
}

func init() {
	RootCmd.AddCommand(bridgeCmd)

	bridgeCmd.Flags().BoolP("ssh-agent", "A", false, "Enable SSH-Agent relay. Sets SSH_SOCK_AUTH in the subprocess call.")
	bridgeCmd.Flags().IntP("listen", "l", 0, "Listen for local queries on specified `port`. The client service responds to http://127.0.0.1:[port]/secrets/key, and outputs the plain text secrets. 'b64:' and 'b64u:' prefixes will provide an encoded value.")

}
