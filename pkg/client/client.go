package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/abourget/secrets-bridge/pkg/bridge"
)

func NewClient(conf *bridge.Bridge) *Client {
	tr := &http.Transport{
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
	}
	return &Client{
		conf: conf,
		httpClient: &http.Client{
			Transport: tr,
		},
		httpTransport: tr,
	}
}

type Client struct {
	conf           *bridge.Bridge
	chosenEndpoint *url.URL
	httpClient     *http.Client
	httpTransport  *http.Transport
}

func (c *Client) Quit() error {
	_, err := c.doRequest("POST", "/quit")
	return err
}

func (c *Client) Ping() error {
	resp, err := c.doRequest("GET", "/ping")
	if string(resp) != "v1" {
		// report better, check if it starts with v\d+ ? indicate the
		// discrepencies between server and client protocol versions.
		return fmt.Errorf(`ping failed: expected protocol version "v1" received %q`, string(resp))
	}
	return err
}

func (c *Client) GetSecret(key string) ([]byte, error) {
	return c.doRequest("GET", fmt.Sprintf("/secrets/%s", key))
}

func (c *Client) GetSecretString(key string) (string, error) {
	resp, err := c.GetSecret(key)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) RequestProxier() func(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(c.chosenEndpoint)
	proxy.Transport = c.httpTransport
	return proxy.ServeHTTP
}

func (c *Client) SSHAgentWebsocketURL() string {
	if c.chosenEndpoint == nil {
		return "https://please-call-ChooseEndpoint-first.../"
	}
	return c.chosenEndpoint.String() + "/ssh-agent-forwarder"
}

func (c *Client) ChooseEndpoint() (err error) {
	chosenEndpoint := make(chan *url.URL, len(c.conf.Endpoints))
	wg := sync.WaitGroup{}
	for _, endpoint := range c.conf.Endpoints {
		wg.Add(1)
		endpoint := endpoint
		go func() {
			defer wg.Done()

			dest := fmt.Sprintf("%s/ping", endpoint)
			req, err := http.NewRequest("GET", dest, nil)
			if err != nil {
				//fmt.Println("ChoosenEndpoint NewRequest:", err)
				return
			}
			resp, err := c.httpClient.Do(req)
			if err != nil {
				//fmt.Println("ChoosenEndpoint Do request:", err)
				return
			}
			resp.Body.Close()

			target, err := url.Parse(endpoint)
			if err != nil {
				log.Println("Error in URL format in endpoints list:", err)
				return
			}

			chosenEndpoint <- target
		}()

	}
	go func() {
		wg.Wait()
		chosenEndpoint <- nil
	}()

	select {
	case recv := <-chosenEndpoint:
		if recv == nil {
			return fmt.Errorf("no valid endpoints found, tried: %q", c.conf.Endpoints)
		}
		c.chosenEndpoint = recv
	}

	return nil

}

func (c *Client) doRequest(method string, path string) ([]byte, error) {
	if c.chosenEndpoint == nil {
		return nil, fmt.Errorf("endpoint not configured, have you called ChooseEndpoint() first ?")
	}

	dest := c.chosenEndpoint.String() + path
	req, err := http.NewRequest(method, dest, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cnt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(cnt))
	}

	return cnt, nil
}
