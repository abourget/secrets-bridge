package agentfwd

import "net"

func DialAgent() (net.Conn, error) {
	return net.Dial("unix", unixSocket)
}
