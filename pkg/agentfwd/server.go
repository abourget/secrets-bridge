package agentfwd

import (
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/net/websocket"
)

var unixSocket = os.Getenv("SSH_AUTH_SOCK")

func TestSSHAgentConnectivity() error {
	client, err := net.Dial("unix", unixSocket)
	if err != nil {
		return err
	}
	return client.Close()
}

func HandleSSHAgentForward(ws *websocket.Conn) {
	fmt.Println("Serving SSH-Agent forward")
	client, err := net.Dial("unix", unixSocket)
	if err != nil {
		log.Println("Can't connect to SSH-Agent:", err)
		return
	}
	defer client.Close()
	defer ws.Close()
	chDone := make(chan bool)

	go func() {
		ioWSCopy(client, ws)
		chDone <- true
	}()
	go func() {
		ioCopyWS(ws, client)
		chDone <- true
	}()
	<-chDone
}
