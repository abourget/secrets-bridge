package bridge

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// NewBridge generates all that is needed to serve a bridge. It generates crypto material (ca cert+key and client cert+key), creates the listener, lists the available IPs.
func NewBridge() (bridge *Bridge, err error) {
	bridge = &Bridge{}

	// Gather IPs
	ips, err := GetAllIPs()
	if err != nil {
		return
	}

	// Setup Listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return
	}

	bridge.Listener = listener
	for _, ip := range ips {
		bridge.Endpoints = append(bridge.Endpoints, fmt.Sprintf("https://%s:%d", ip.String(), listener.Addr().(*net.TCPAddr).Port))
	}

	// Generate CA key+cert
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(01),
		Subject:               pkix.Name{Organization: []string{"secrets-bridge"}, CommonName: "secrets-bridge"},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		BasicConstraintsValid: true,
	}
	tmpl.IsCA = true
	tmpl.MaxPathLenZero = true
	tmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	tmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	tmpl.IPAddresses = ips

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &privKey.PublicKey, privKey)
	if err != nil {
		return
	}

	caCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	bridge.CACert = string(caCert)

	caKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	caKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: caKeyBytes})

	bridge.caTLSCert, err = tls.X509KeyPair([]byte(caCert), caKey)
	if err != nil {
		return
	}

	// Generate client key + csr + cert

	return
}
