package websocket

import (
	"fmt"
	"net/http"
	"net/url"
)

type WebSocketClient struct {
	client http.Client
}

func (c *WebSocketClient) handshake(rawURL string) error {
	
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println(err)
	}
	
	h := &http.Header{}
	h.Add("Host", parsedURL.Host)
	h.Add("Upgrade", "websocket")
	h.Add("Connection", "Upgrade")
	h.Add("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	
	req := &http.Request{
		Method: "GET",
		Header: *h,
		URL:    parsedURL,
	}

	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)

	return nil
}

func (wc *WebSocketClient) Connect(url string) {
	wc.handshake(url)
}
