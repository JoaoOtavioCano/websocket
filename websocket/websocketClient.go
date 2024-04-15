package websocket

import (
	"crypto/rand"
	"encoding/base64"
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
		return err
	}

	websocketKey, err := createWebsocketKeyValue()
	if err != nil {
		return nil
	}

	h := &http.Header{}
	h.Add("Host", parsedURL.Host)
	h.Add("Upgrade", "websocket")
	h.Add("Connection", "Upgrade")
	h.Add("Sec-WebSocket-Key", websocketKey)
	h.Add("Sec-WebSocket-Version", "13")

	req := &http.Request{
		Method: "GET",
		Header: *h,
		URL:    parsedURL,
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	checkStatusCode := resp.StatusCode == http.StatusSwitchingProtocols
	checkUpgradeHeader := len(resp.Header[http.CanonicalHeaderKey("Upgrade")]) == 1 && resp.Header[http.CanonicalHeaderKey("Upgrade")][0] == "websocket"
	checkConnectionHeader := len(resp.Header[http.CanonicalHeaderKey("Connection")]) == 1 && resp.Header[http.CanonicalHeaderKey("Connection")][0] == "Upgrade"
	checkWebsocketAcceptHeader := len(resp.Header[http.CanonicalHeaderKey("Sec-WebSocket-Accept")]) == 1 && resp.Header[http.CanonicalHeaderKey("Sec-WebSocket-Accept")][0] == CreateWebsocketAcceptValue(websocketKey)


	if !checkStatusCode || !checkUpgradeHeader || !checkConnectionHeader || !checkWebsocketAcceptHeader {
		fmt.Println("handshake error: connection not established")
		return fmt.Errorf("handshake error: connection not established")
	}

	return nil
}

func (wc *WebSocketClient) Connect(url string) {
	wc.handshake(url)
}

func createWebsocketKeyValue() (string, error) {
	nonce := make([]byte, 16)

	_, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}

	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)

	return nonceBase64, nil
}
