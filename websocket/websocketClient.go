package websocket

import (
	"net/http"
	"net/url"
)

type WebSocketClient struct {
	client http.Client
}

func (c *WebSocketClient) handshake(host string) error {
	h := &http.Header{}
	h.Add("Upgrade", "websocket")
	h.Add("Connection", "Upgrade")
	h.Add("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	url := &url.URL{Host: host}

	req := &http.Request{
		Method: "GET",
		Header: *h,
		URL:    url,
	}

	_, err := c.client.Do(req)
	if err != nil {
		return err
	}

	
}

func (wc *WebSocketClient) Connect() {
	wc.handshake("http://localhost:8000/")
}
