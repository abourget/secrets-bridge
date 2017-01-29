package bridge

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
)

type Bridge struct {
	Endpoints []string `json:"endpoints"`

	CACert     string `json:"ca_cert"`
	caCertPool *x509.CertPool
	caKey      []byte `json:"-"`
	caTLSCert  tls.Certificate

	ClientCert    string `json:"client_cert"`
	ClientKey     string `json:"client_key"`
	clientTLSCert tls.Certificate

	Listener net.Listener `json:"-"`
}

func NewFromString(b64conf string) (bridge *Bridge, err error) {
	content, err := base64.StdEncoding.DecodeString(b64conf)
	if err != nil {
		return nil, fmt.Errorf("decoding base64 input: %s", err)
	}

	err = json.Unmarshal(content, &bridge)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal: %s", err)
	}

	bridge.caCertPool, err = bridge.readCACertPool()
	if err != nil {
		return nil, fmt.Errorf("building CA cert pool: %s", err)
	}

	bridge.clientTLSCert, err = tls.X509KeyPair([]byte(bridge.ClientCert), []byte(bridge.ClientKey))
	if err != nil {
		return nil, fmt.Errorf("loading client keypair: %s", err)
	}

	return
}

func (b *Bridge) readCACertPool() (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()

	block, _ := pem.Decode([]byte(b.CACert))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM encoding for ca_cert field")
	}
	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, fmt.Errorf("ca_cert should have a single CERTIFICATE block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	caCertPool.AddCert(cert)

	return caCertPool, nil
}

func (b *Bridge) ClientTLSConfig() *tls.Config {
	return &tls.Config{
		RootCAs:      b.caCertPool,
		Certificates: []tls.Certificate{b.clientTLSCert}, // populated through `NewFromString`
	}
}

func (b *Bridge) ServerTLSConfig() *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{b.caTLSCert}, // populated through `NewBridge`
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}
