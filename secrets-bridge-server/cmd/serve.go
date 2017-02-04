package cmd

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/secrets-bridge/pkg/agentfwd"
	"github.com/abourget/secrets-bridge/pkg/bridge"
	"github.com/abourget/secrets-bridge/pkg/secrets"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
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
var insecureMode bool

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().BoolVarP(&enableSSHAgent, "ssh-agent-forwarder", "A", false, "Enable SSH Agent forwarder. Uses env's SSH_AUTH_SOCK.")
	serveCmd.Flags().StringSliceVar(&secretLiterals, "secret", []string{}, "Literal secret, in the form `key=value`. 'key' can be prefixed by 'b64:' or 'b64u:' to denote that the 'value' is base64-encoded or base64-url-encoded")
	serveCmd.Flags().StringSliceVar(&secretsFromFiles, "secret-from-file", []string{}, "Secret from the content of a file, in the form `key=filename`. 'key' can also be prefixed by 'b64:' and 'b64u:' to indicate the encoding of the file")
	serveCmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Timeout in `seconds` before the server exits. Defaults to 0 (indefinite)")
	serveCmd.Flags().BoolVarP(&insecureMode, "insecure", "", false, "Do not check client certificate for incoming connections")
}

func serve(cmd *cobra.Command, args []string) {
	b, err := bridge.NewBridge()
	if err != nil {
		log.Fatalln("Failed to setup bridge:", err)
	}

	jsonConfig, _ := json.Marshal(b)

	log.Printf("Writing %q\n", bridgeConfFilename)
	err = ioutil.WriteFile(bridgeConfFilename, []byte(base64.StdEncoding.EncodeToString(jsonConfig)), 0600)
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

		err = store.Add(parts[0], fileContent)
		if err != nil {
			log.Fatalf(`Error reading value from secrets file %q: %s\n`, filename, err)
		}
	}
	log.Printf("Loaded %d secrets\n", len(store.Secrets))

	mux := http.NewServeMux()
	mux.HandleFunc("/secrets/", func(w http.ResponseWriter, r *http.Request) {
		matches := secretsRE.FindStringSubmatch(r.URL.Path)
		if matches == nil {
			http.NotFound(w, r)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := matches[1]
		value := store.Get(key)
		if value == nil {
			http.NotFound(w, r)
			return
		}

		log.Printf("Serving secret %q (%d bytes)\n", key, len(value))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(value)))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(200)
		w.Write(value)
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a PING, sending protocol version.")
		w.Write([]byte("v1"))
	})
	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received QUIT, quitting...")
		w.Write([]byte("quitting..."))
		go func() {
			time.Sleep(10 * time.Millisecond)
			os.Exit(0)
		}()
	})
	if enableSSHAgent {
		log.Println("Enabling SSH-Agent forwarding handler")
		mux.Handle("/ssh-agent-forwarder", websocket.Handler(agentfwd.HandleSSHAgentForward))
	}

	server := http.Server{
		Handler: mux,
	}
	tlsConfig := b.ServerTLSConfig(insecureMode)
	tlsListener := tls.NewListener(b.Listener, tlsConfig)

	done := make(chan bool, 2)
	go func() {
		log.Println("Serving requests")
		err = server.Serve(tlsListener)
		if err != nil {
			log.Fatalln("error serving requests:", err)
		}
		log.Println("Gracefully exiting")
		done <- true
	}()

	if timeout != 0 {
		go func() {
			<-time.After(time.Duration(timeout) * time.Second)
			log.Fatalf("Server shutting down after timeout of %d seconds\n", timeout)
			done <- false
		}()
	}

	<-done
}

var secretsRE = regexp.MustCompile(`/secrets/(.+)`)
