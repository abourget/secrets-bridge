// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bridgeCmd represents the bridge command
var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		// TODO: validate args isn't empty, as we'll call that command...
		// connect using the client,
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

		fmt.Println("bridge called")
	},
}

func init() {
	RootCmd.AddCommand(bridgeCmd)

	bridgeCmd.Flags().BoolP("ssh-agent", "A", false, "Enable SSH-Agent relay. Sets SSH_SOCK_AUTH in the subprocess call.")
	bridgeCmd.Flags().IntP("listen", "l", 0, "Listen for local queries on specified `port`. The client service responds to http://127.0.0.1:[port]/secrets/key, and outputs the plain text secrets. 'b64:' and 'b64u:' prefixes will provide an encoded value.")

}
