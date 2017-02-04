package agentfwd

import "os"

var unixSocket = os.Getenv("SSH_AUTH_SOCK")
