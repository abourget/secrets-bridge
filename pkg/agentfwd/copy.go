package agentfwd

import (
	"io"

	"golang.org/x/net/websocket"
)

// io_copy_ws and io_ws_copy both taken from: https://github.com/influxdata/surgemq/blob/master/examples/surgemq/websocket.go

/* copy from websocket to writer, this copies the binary frames as is */
func ioCopyWS(src *websocket.Conn, dst io.Writer) (int, error) {
	var buffer []byte
	count := 0
	for {
		err := websocket.Message.Receive(src, &buffer)
		if err != nil {
			return count, err
		}
		n := len(buffer)
		count += n
		i, err := dst.Write(buffer)
		if err != nil || i < 1 {
			return count, err
		}
	}
	//return count, nil
}

/* copy from reader to websocket, this copies the binary frames as is */
func ioWSCopy(src io.Reader, dst *websocket.Conn) (int, error) {
	buffer := make([]byte, 2048)
	count := 0
	for {
		n, err := src.Read(buffer)
		if err != nil || n < 1 {
			return count, err
		}
		count += n
		err = websocket.Message.Send(dst, buffer[0:n])
		if err != nil {
			return count, err
		}
	}
	//return count, nil
}
