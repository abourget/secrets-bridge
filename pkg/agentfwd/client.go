package agentfwd

import (
	"crypto/tls"
	"log"
	"net"
	"os"

	"golang.org/x/net/websocket"
)

func ListenAndServeSSHAgentForwarder(targetServer, websocketOrigin string, tlsConfig *tls.Config) error {
	unixSocket := "/tmp/secrets-bridge-ssh-agent-forwarder"
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: unixSocket, Net: "unix"})
	if err != nil {
		return err
	}
	defer os.Remove(unixSocket)

	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return err
		}

		go func() {
			defer conn.Close()

			config, err := websocket.NewConfig(targetServer, websocketOrigin)
			if err != nil {
				log.Println("couldn't configure websocket connection:", err)
				return
			}

			config.TlsConfig = tlsConfig

			ws, err := websocket.DialConfig(config)
			if err != nil {
				log.Println("couldn't connect to websocket:", err)
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
