package agentfwd

import (
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

func TestSSHAgentConnectivity() error {
	client, err := DialAgent()
	if err != nil {
		return err
	}
	return client.Close()
}

func HandleSSHAgentForward(ws *websocket.Conn) {
	fmt.Println("Serving SSH-Agent forward")
	client, err := DialAgent()
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
