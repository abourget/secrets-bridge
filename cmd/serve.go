package cmd

import (
	"bytes"
	"compress/gzip"
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
	Short: "Serves an SSH Agent forwarder over the network, and secrets",
	Long:  ``,
	Run:   serveDaemonized,
}

var caKeyStore string
var secretLiterals []string
var secretsFromFiles []string
var enableSSHAgent bool
var timeout int
var insecureMode bool
var writeConf bool
var daemonize string

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&bridgeConfFilename, "bridge-conf-file", "f", "", "Bridge authentication file. Written to in serve command, read in kill command. Defaults to `~/.bridge-conf`")
	serveCmd.Flags().BoolVarP(&writeConf, "write-conf", "w", false, "Write the bridge config to a file instead of printing out. Specify file with '--bridge-conf-file'.")

	serveCmd.Flags().StringVarP(&caKeyStore, "ca-key-store", "", "", "Filenam where to read/store the CA Key if you want to reuse, to avoid changing the bridge conf, thus avoiding Docker rebuilds.")
	serveCmd.Flags().BoolVarP(&enableSSHAgent, "ssh-agent-forwarder", "A", false, "Enable SSH Agent forwarder. Uses env's SSH_AUTH_SOCK.")
	serveCmd.Flags().StringVarP(&daemonize, "daemonize", "d", "", "Daemonize after listening socket successfully opened. The parameter is the output file to log stdout / stderr.")
	serveCmd.Flags().StringSliceVar(&secretLiterals, "secret", []string{}, "Literal secret, in the form `key=value`. 'key' can be prefixed by 'b64:' or 'b64u:' to denote that the 'value' is base64-encoded or base64-url-encoded")
	serveCmd.Flags().StringSliceVar(&secretsFromFiles, "secret-from-file", []string{}, "Secret from the content of a file, in the form `key=filename`. 'key' can also be prefixed by 'b64:' and 'b64u:' to indicate the encoding of the file")
	serveCmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Timeout in `seconds` before the server exits. Defaults to 0 (indefinite)")
	serveCmd.Flags().BoolVarP(&insecureMode, "insecure", "", false, "Do not check client certificate for incoming connections")
}

func serve(cmd *cobra.Command, args []string) {
	confFile := bridgeConfFilenameWithDefault()
	var b *bridge.Bridge
	var err error
	if caKeyStore != "" {
		b, err = bridge.NewCachedBridge(caKeyStore, confFile)
		if err != nil {
			log.Println("WARNING: couldn't load configuration from provided --ca-key-store:", err)
			b = nil
		}
	}

	if b == nil {
		b, err = bridge.NewBridge(caKeyStore)
		if err != nil {
			log.Fatalln("Failed to setup bridge:", err)
		}

		jsonConfig, _ := json.Marshal(b)

		buf := &bytes.Buffer{}

		gz, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
		gz.Write([]byte(jsonConfig))
		gz.Close()

		configBytes := buf.Bytes()

		bridgeConfText := base64.RawURLEncoding.EncodeToString(configBytes)
		if writeConf {
			log.Printf("Writing bridge conf to %q\n", confFile)

			err = ioutil.WriteFile(confFile, []byte(bridgeConfText), 0600)
			if err != nil {
				log.Fatalf("Error writing %q: %s\n", confFile, err)
			}
		} else {
			log.Printf("Bridge config: %s\n", bridgeConfText)
		}
	}

	// Read secrets
	store := &secrets.Store{}
	for _, secret := range secretLiterals {
		parts := strings.SplitN(secret, "=", 2)
		if len(parts) != 2 {
			log.Fatalf(`Invalid secret literal, expected format "key=filename", or "filename" got %q\n`, secret)
		}
		err := store.Add(parts[0], []byte(parts[1]))
		if err != nil {
			log.Fatalf(`Error reading literal secret %q: %s"`, secret, err)
		}
	}

	for _, secret := range secretsFromFiles {
		parts := strings.SplitN(secret, "=", 2)
		if len(parts) > 2 {
			log.Fatalf(`Invalid secret from file, expected format "key=filename", got %q\n`, secret)
		}

		var filename string
		if len(parts) == 1 {
			filename = parts[0]
		} else {
			filename = parts[1]
		}
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
	} else {
		log.Println("SSH-Agent forwarder IS NOT ENABLED. Use -A to enable it.")
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

	detachFromParent()

	<-done
}

var secretsRE = regexp.MustCompile(`/secrets/(.+)`)
