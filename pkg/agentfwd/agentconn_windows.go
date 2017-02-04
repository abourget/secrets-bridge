package agentfwd

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

func DialAgent() (conn net.Conn, err error) {
	log.Println("Using the unix socket as a file, unwrapping cygwin/msysgit's fake unix domain socket")
	cnt, err := ioutil.ReadFile(unixSocket)
	if err != nil {
		return nil, fmt.Errorf("opening %q: %s", unixSocket, err)
	}

	matches := windowsFakeSocket.FindStringSubmatch(string(cnt))
	if matches == nil {
		return nil, fmt.Errorf("couldn't read SSH_AUTH_SOCK file %q as a fake unix socket", unixSocket)
	}

	tcpPort := matches[1]
	isCygwin := matches[2] == "s " // as opposed to msysgit
	key := matches[3]

	uid, err := extractMSysGitUID()
	if err != nil {
		return
	}

	conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%s", tcpPort))
	if err != nil {
		return
	}

	if isCygwin {
		b := make([]byte, 16)
		fmt.Sscanf(key,
			"%02x%02x%02x%02x-%02x%02x%02x%02x-%02x%02x%02x%02x-%02x%02x%02x%02x",
			&b[3], &b[2], &b[1], &b[0],
			&b[7], &b[6], &b[5], &b[4],
			&b[11], &b[10], &b[9], &b[8],
			&b[15], &b[14], &b[13], &b[12],
		)

		// fmt.Println("Writing first GUID bytes")
		if _, err = conn.Write(b); err != nil {
			return nil, fmt.Errorf("write b: %s", err)
		}

		// fmt.Println("Reading b2")
		b2 := make([]byte, 16)
		if _, err = conn.Read(b2); err != nil {
			return nil, fmt.Errorf("read b2: %s", err)
		}
		// fmt.Printf("Received b2: %q %s\n", b2, string(b2))

		// fmt.Println("Writing pid,gid,uid")
		pidsUids := make([]byte, 12)
		pid := os.Getpid()
		gid := pid // for cygwin's AF_UNIX -> AF_INET, pid = gid
		binary.LittleEndian.PutUint32(pidsUids, uint32(pid))
		binary.LittleEndian.PutUint32(pidsUids[4:], uint32(uid))
		binary.LittleEndian.PutUint32(pidsUids[8:], uint32(gid))
		// fmt.Println("  Writing", pidsUids, string(pidsUids))
		if _, err = conn.Write(pidsUids); err != nil {
			return nil, fmt.Errorf("write pid,uid,gid: %s", err)
		}

		// fmt.Println("Reading b3")
		b3 := make([]byte, 12)
		if _, err = conn.Read(b3); err != nil {
			return nil, fmt.Errorf("read pid,uid,gid: %s", err)
		}
		// fmt.Printf("Received b3: %v %s\n", b3, string(b3))

	} else {
		// We should implement the last bit of this page: http://stackoverflow.com/questions/23086038/what-mechanism-is-used-by-msys-cygwin-to-emulate-unix-domain-sockets
		return nil, fmt.Errorf("MSysGit ssh-agent implementation not yet supported.")
	}

	// fmt.Println("Good! Continuing...")
	return conn, nil
}

// implements http://stackoverflow.com/questions/23086038/what-mechanism-is-used-by-msys-cygwin-to-emulate-unix-domain-sockets
// counterpart to: https://cygwin.com/git/gitweb.cgi?p=newlib-cygwin.git;a=blob;f=winsup/cygwin/net.cc;h=fd903b1124178f34cec12a361b7be800caf3ead6;hb=HEAD#l919

var windowsFakeSocket = regexp.MustCompile(`!<socket >(\d+) (s )?([A-Fa-f0-9-]+)`)

func extractMSysGitUID() (uid int64, err error) {
	out, err := exec.Command("bash.exe", "-c", "ps").Output()
	fmt.Println("out", string(out))
	if err != nil {
		return 0, err
	}

	matches := myPSLine.FindStringSubmatch(string(out))
	if matches == nil {
		return 0, fmt.Errorf("couldn't find unix-like UID while running 'bash -c ps'")
	}

	return strconv.ParseInt(matches[1], 10, 32)
}

var myPSLine = regexp.MustCompile(`(?m)^\s+\d+\s+\d+\s+\d+\s+\d+\s+\?\s+(\d+)`)
