package agentfwd

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

var UnixSocket = "/tmp/secrets-bridge-ssh-agent-forwarder"

func ListenAndServeSSHAgentForwarder(targetURL, websocketOrigin string, tlsConfig *tls.Config) error {
	os.Remove(UnixSocket)
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: UnixSocket, Net: "unix"})
	if err != nil {
		return err
	}
	defer os.Remove(UnixSocket)

	targetURL = strings.Replace(targetURL, "https:", "wss:", 1)

	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return err
		}

		log.Println("secrets-bridge: Received connection on SSH-Agent unix socket")

		go func() {
			defer conn.Close()

			config, err := websocket.NewConfig(targetURL, websocketOrigin)
			if err != nil {
				log.Println("secrets-bridge: couldn't configure websocket connection:", err)
				return
			}

			config.TlsConfig = tlsConfig

			ws, err := websocket.DialConfig(config)
			if err != nil {
				log.Println("secrets-bridge: ssh-forward failed (is it started on server?):", err)
				return
			}
			defer ws.Close()

			chDone := make(chan bool)
			go func() {
				ioWSCopy(conn, ws)
				chDone <- true
			}()
			go func() {
				ioCopyWS(ws, conn)
				chDone <- true
			}()
			<-chDone
		}()
	}
}
