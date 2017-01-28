package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/secrets"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: serve,
}

var secretLiterals []string
var secretsFromFiles []string
var enableSSHAgent bool
var timeout int

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().BoolVarP(&enableSSHAgent, "ssh-agent-forwarder", "A", false, "Enable SSH Agent forwarder. Uses env's SSH_SOCK_AUTH.")
	serveCmd.Flags().StringSliceVar(&secretLiterals, "secret", []string{}, "Literal secret, in the form `key=value`. 'key' can be prefixed by 'b64:' or 'b64u:' to denote that the 'value' is base64-encoded or base64-url-encoded")
	serveCmd.Flags().StringSliceVar(&secretsFromFiles, "secret-from-file", []string{}, "Secret from the content of a file, in the form `key=filename`. 'key' can also be prefixed by 'b64:' and 'b64u:' to indicate the encoding of the file")
	serveCmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Timeout in `seconds` before the server exits. Defaults to 0 (indefinite)")
}

func serve(cmd *cobra.Command, args []string) {

	// TODO: Work your own magic here
	fmt.Println("serve called")
	b, err := bridge.NewBridge()
	if err != nil {
		log.Fatalln("Failed to setup bridge:", err)
	}

	bridgeConfFilename, err := cmd.Flags().GetString("bridge-conf")
	if err != nil {
		bridgeConfFilename = "bridge-conf"
	}

	jsonConfig, _ := json.Marshal(b)
	err = ioutil.WriteFile(bridgeConfFilename, jsonConfig, 0600)
	if err != nil {
		log.Fatalf("Error writing %q: %s\n", bridgeConfFilename, err)
	}

	// Read secrets
	store := &secrets.Store{}

	for _, secret := range secretLiterals {
		parts := strings.SplitN(secret, "=", 2)
		if len(parts) != 2 {
			log.Fatalf(`Invalid secret literal, expected format "key=value", got %q\n`, secret)
		}
		err := store.Add(parts[0], []byte(parts[1]))
		if err != nil {
			log.Fatalf(`Error reading literal secret %q: %s"`, secret, err)
		}
	}

	for _, secret := range secretsFromFiles {
		parts := strings.SplitN(secret, "=", 2)
		if len(parts) != 2 {
			log.Fatalf(`Invalid secret from file, expected format "key=filename", got %q\n`, secret)
		}

		filename := parts[1]
		fileContent, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf(`Error reading file %q for secret-from-file %q: %s\n`, filename, secret, err)
		}

		err := store.Add(parts[0], fileContent)
		if err != nil {
			log.Fatalf(`Error reading value from secrets file %q: %s\n`, filename, err)
		}
	}

	// depending on `-A`, setup the websocket endpoint.
	// generate the key meterial, new client cert/key pair, CA, sign it, etc..
	// the listener should be TLS configured, with those certs, and only accept connections
	//   using this client-cert thing..
	//
	// handle websocket connections, and forward between that and
	// the unix domain socket we connected to.

	// handle secrets requests, which will be merely forwarded,
	// and b64 things will be asked right here.. we'll merely
	// forward from the client when he listen on `--listen`.. both
	// status codes and content.
	//
	// handler("GET", "/secrets/:key")
	// handler("POST", "/ssh-agent-forwarder")

	// listen for a `--timeout` and exit on timeout.. or not..
}
