package bridge

import "net"

func GetAllIPs() (out []net.IP, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return out, err
		}

		for _, addr := range addrs {
			switch ip := addr.(type) {
			case *net.IPNet:
				out = append(out, ip.IP)
			case *net.IPAddr:
				out = append(out, ip.IP)
			}
		}
	}

	return
}
