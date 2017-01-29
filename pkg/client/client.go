package client

import (
	"net"
	"net/http"
	"time"

	"github.com/abourget/secrets-bridge/pkg/bridge"
)

func NewClient(conf *bridge.Bridge) *Client {
	return &Client{
		conf: conf,
		httpClient: &http.Client{
			Transport: &http.Transport{
				// similar to `http.DefaultTransport`, with additional `TLSConfig`.
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig:       conf.ClientTLSConfig(),
			},
		},
	}
}

type Client struct {
	conf           *bridge.Bridge
	chosenEndpoint string
	httpClient     *http.Client
}

func (c *Client) Quit() error {
	_, err := c.doRequest("POST", "/quit")
	return err
}

func (c *Client) Ping() error {
	// TODO: loop through the addresses, when we find one that works,
	// we reconfigure the Client's `auth` field, with the first
	// endpoint being the one that worked, so next calls we don't need
	// to cycle through all of them.

	// TODO: ping would check the return value too, make sure it's
	// "pong" and a 200 status code.. so we know we're talking to the
	// right beast.

	// TODO: ping should fail if the version number negotiated isn't
	// compatible with this version protocol, in which case the client
	// spits out an error, saying "protocols incompatible, upgrade
	// either your client or server", report server protocol version,
	// and client protocol version.
	_, err := c.doRequest("GET", "/ping")
	return err
}

func (c *Client) doRequest(method string, path string) (string, error) {
	// TODO: do the HTTPS call, adding the Server Cert, and the Client
	// cert/key pair to the TLSConfig.
	// perform the request
	// get the response
	// return on errors..
	return "", nil
}
