package agentfwd

import (
	"fmt"
	"log"

	"github.com/xanzy/ssh-agent"
	"golang.org/x/net/websocket"
)

func TestSSHAgentConnectivity() (err error) {
	_, client, err := sshagent.New()
	if err != nil {
		log.Println("Can't connect to SSH-Agent:", err)
		return
	}
	client.Close()

	return nil
}

func HandleSSHAgentForward(ws *websocket.Conn) {
	fmt.Println("Serving SSH-Agent forward")
	_, client, err := sshagent.New()
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
