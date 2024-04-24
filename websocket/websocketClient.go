package websocket

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

type WebSocketClient struct {
	client http.Client
	conn   net.TCPConn
}

func (c *WebSocketClient) Handshake() error {

	rawURL := fmt.Sprintf("http://%s", c.conn.RemoteAddr().String())

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
		Close:  false,
	}

	req.Write(&c.conn)

	buf := make([]byte, 4096)

	for {
		_, err := c.conn.Read(buf)
		if err != nil {
			fmt.Printf("Error reading from socket: %v\n", err)
			break
		}

		break
	}

	resp := &HandshakeResponse{}

	resp.Decode(buf)

	checkStatusCode := resp.statusCode == http.StatusSwitchingProtocols
	checkUpgradeHeader := resp.upgrade != "" && resp.upgrade == "websocket"
	checkConnectionHeader := resp.connection != "" && resp.connection == "Upgrade"
	checkWebsocketAcceptHeader := resp.secWebsocketAccept != "" && resp.secWebsocketAccept == CreateWebsocketAcceptValue(websocketKey)

	if !checkStatusCode || !checkUpgradeHeader || !checkConnectionHeader || !checkWebsocketAcceptHeader {
		fmt.Println("handshake error: connection not established")
		return fmt.Errorf("handshake error: connection not established")
	}

	return nil
}

func NewWebSocketClient(url string) (*WebSocketClient, error) {
	add, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, add)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	c := &http.Client{}

	return &WebSocketClient{
		conn:   *conn,
		client: *c,
	}, nil
}

func (c *WebSocketClient) Write(message string) {
	data := []byte(message)
	_, err := c.conn.Write(data)
	if err != nil {
		fmt.Println(err)
	}
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
