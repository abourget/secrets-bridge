package bridge

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"strings"
	"time"
)

func NewCachedBridge(caKeyStore, confFile string) (bridge *Bridge, err error) {
	bridgeConf, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}

	bridge, err = NewFromString(string(bridgeConf))
	if err != nil {
		return
	}

	if len(bridge.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints listed in cached bridge conf, can't listen on the same port")
	}

	firstEndpoint := bridge.Endpoints[0]
	segments := strings.Split(firstEndpoint, ":")
	listenPort := segments[len(segments)-1]

	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		return
	}

	bridge.Listener = listener

	caKey, err := ioutil.ReadFile(caKeyStore)
	if err != nil {
		return
	}

	bridge.caTLSCert, err = tls.X509KeyPair([]byte(bridge.CACert), caKey)
	if err != nil {
		return
	}

	return
}

// NewBridge generates all that is needed to serve a bridge. It generates crypto material (ca cert+key and client cert+key), creates the listener, lists the available IPs.
func NewBridge(caKeyStore string) (bridge *Bridge, err error) {
	bridge = &Bridge{}

	ips, err := GetAllIPs()
	if err != nil {
		return
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return
	}

	bridge.Listener = listener
	for _, ip := range ips {
		ipStr := ip.String()
		if strings.Contains(ipStr, ":") {
			ipStr = "[" + ipStr + "]"
		}
		bridge.Endpoints = append(bridge.Endpoints, fmt.Sprintf("https://%s:%d", ipStr, listener.Addr().(*net.TCPAddr).Port))
	}

	// Generate CA key+cert
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	caCertTpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"secrets-bridge"},
			CommonName:   "secrets-bridge-server",
		},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		BasicConstraintsValid: true,
		IsCA:           true,
		MaxPathLenZero: true,
		KeyUsage:       x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IPAddresses:    ips,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, caCertTpl, caCertTpl, &privKey.PublicKey, privKey)
	if err != nil {
		return
	}

	bridge.CACert = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}))

	bridge.caCertPool, err = bridge.readCACertPool()
	if err != nil {
		return
	}

	caKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	if caKeyStore != "" {
		log.Println("Writing ca key store", caKeyStore)
		if err = ioutil.WriteFile(caKeyStore, []byte(caKey), 0600); err != nil {
			return
		}
	}

	bridge.caTLSCert, err = tls.X509KeyPair([]byte(bridge.CACert), caKey)
	if err != nil {
		return
	}

	// Generate client key + csr + cert
	clientCertTpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"secrets-bridge"},
			CommonName:   "secrets-bridge",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(1 * time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	clientPriv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return
	}

	clientCert, err := x509.CreateCertificate(rand.Reader, clientCertTpl, caCertTpl, &clientPriv.PublicKey, privKey)
	if err != nil {
		return
	}

	bridge.ClientCert = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCert,
	}))

	// if ok := bridge.caCertPool.AppendCertsFromPEM([]byte(bridge.ClientCert)); !ok {
	// 	return nil, fmt.Errorf("oh mama")
	// }

	bridge.ClientKey = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientPriv),
	}))

	return
}
