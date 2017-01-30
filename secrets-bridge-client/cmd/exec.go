// Copyright © 2017 Alexandre Bourget <alex@bourget.cc>

package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/abourget/secrets-bridge/pkg/agentfwd"
	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/client"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Setup bridge, connect to server and proxy the SSH-Agent or serve secrets securely.",
	Long: `Example:

secrets-bridge-client --bridge-conf=$BRIDGE_CONF exec --listen 4444 -- ./command-that-will-curl-from-port-4444.sh

secrets-bridge-client --bridge-conf=$BRIDGE_CONF exec
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("no command to execute specified, see examples with --help")
		}

		bridgeConf, err := cmd.Flags().GetString("bridge-conf")
		if err != nil {
			log.Fatalln("--bridge-conf invalid:", err)
		}

		bridge, err := bridge.NewFromString(bridgeConf)
		if err != nil {
			log.Fatalln("--bridge-conf has an invalid value:", err)
		}

		c := client.NewClient(bridge)

		err = c.ChooseEndpoint()
		if err != nil {
			log.Fatalln("error pinging server:", err)
		}

		log.Println("secrets-bridge: server connection successful")

		// setup the SSH agent if we specified `--ssh-agent`, with websocket and all..
		// setup the listener BEFORE exec'ing the subprocess
		// exec.Command(args...)
		// quit when we're done, delete the unix socket is we were doing ssh-agent.

		if enableSSHAgentForwarding {
			go func() {
				log.Println("secrets-bridge: Setting up SSH-Agent forwarder...")
				err := agentfwd.ListenAndServeSSHAgentForwarder(c.SSHAgentWebsocketURL(), "https://localhost", bridge.ClientTLSConfig())
				if err != nil {
					log.Fatalln("couldn't setup SSH-Agent forwarder:", err)
				}
			}()
			os.Setenv("SSH_AUTH_SOCK", agentfwd.UnixSocket)
			time.Sleep(25 * time.Millisecond)
		}

		if listenOnPort != 0 {
			http.HandleFunc("/secrets/", c.RequestProxier())
			go func() {
				log.Printf("secrets-bridge: Serving secrets on 127.0.0.1:%d\n", listenOnPort)
				http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", listenOnPort), nil)
			}()
			time.Sleep(25 * time.Millisecond)
		}

		// implement things like --secret-to-file key=filename
		// and wipe the files written to disk after

		program := args[0]
		var arguments []string
		if len(args) > 1 {
			arguments = args[1:]
		}
		subprocess := exec.Command(program, arguments...)
		subprocess.Env = os.Environ()
		subprocess.Stdin = os.Stdin
		subprocess.Stdout = os.Stdout
		subprocess.Stderr = os.Stderr

		log.Printf("secrets-bridge: Calling subprocess with: %q\n", args)
		subErr := subprocess.Run()
		log.Printf("secrets-bridge: Call returned, exited with: %s\n", subErr)

		if enableSSHAgentForwarding {
			os.Remove(agentfwd.UnixSocket)
		}

		// now reflect the subprocess error...
		if subErr != nil {
			os.Exit(199)
		}

		// var failedRemovals bool
		// for _, sec := range secretToFiles {
		// 	parts := strings.SplitN(sec, "=", 2) // we checked for errors earlier
		// 	log.Printf("secrets-bridge: removing %q\n", parts[1])
		// 	err := os.Remove(parts[1])
		// 	if err != nil {
		// 		failedRemovals = true
		// 		log.Printf("error removing %q: %s\n", parts[1], err)
		// 	}
		// }
		// if failedRemovals {
		// 	log.Fatalln("failed to remove some secrets, exiting with error")
		// }
	},
}

var enableSSHAgentForwarding bool
var listenOnPort int
var secretToFiles []string

func init() {
	RootCmd.AddCommand(execCmd)

	execCmd.Flags().BoolVarP(&enableSSHAgentForwarding, "ssh-agent", "A", false, "Enable SSH-Agent relay. Sets SSH_AUTH_SOCK in the subprocess call.")
	execCmd.Flags().IntVarP(&listenOnPort, "listen", "l", 0, "Listen for local queries on specified `port`. The client service responds to http://127.0.0.1:[port]/secrets/key, and outputs the plain text secrets. 'b64:' and 'b64u:' prefixes will provide an encoded value.")
	// execCmd.Flags().StringSliceVarP(&secretToFiles, "secret-to-file", "s", []string{}, "Use `key=output_filename`. Write secret to temporary file before calling the command, and clean-up no matter what happens to the subprocess. This ensures you won't leave traces in a Docker image layer, for example.")

}
