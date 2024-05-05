package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type WebSocketServer struct {
	listener net.TCPListener
	conn     net.Conn
}

type HandshakeRequest struct {
	method              string
	host                string
	connection          string
	secWebsocketKey     string
	secWebsocketVersion string
	upgrade             string
}

type HandshakeResponse struct {
	proto              string
	statusCode         int
	statusText         string
	connection         string
	secWebsocketAccept string
	upgrade            string
}

func NewWebSocketSever(url string) (*WebSocketServer, error) {
	add, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	l, err := net.ListenTCP("tcp", add)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &WebSocketServer{
		listener: *l,
	}, nil
}

func (ws *WebSocketServer) ListenAndServe() {

	fmt.Printf("Listening %s\n", ws.listener.Addr())

	conn, err := ws.listener.Accept()
	if err != nil {
		fmt.Println(err)
	}
	go ws.handshake(conn)

	for {
		if ws.conn != nil {
			break
		}
	}
}

func (ws *WebSocketServer) ReadMessage() ([]byte, error) {

	reader := bufio.NewReader(ws.conn)

	buf := make([]byte, 30)

	reader.Read(buf)

	f, err := ParseFrame(buf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	//fmt.Println(buf)
	fmt.Println(f.encode())

	return f.payloadData, err
}

func (ws *WebSocketServer) Write(data []byte) error {
	f := newFrame(textFrame, false, data)

	_, err := ws.conn.Write(f.encode())
	if err != nil {
		return err
	}

	return nil
}

func (ws *WebSocketServer) handshake(conn net.Conn) {

	buffer := make([]byte, 1024)
	r := &HandshakeRequest{}

	for {
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		r = parseHandshakeRequest(buffer)

		break
	}

	/*
		resp := &HandshakeResponse{}

		checkReqMethod := r.method == "GET"
		checkUpgradeHeader := r.upgrade != "" && strings.Compare(r.upgrade, "websocket") == 0
		checkConnectionHeader := r.connection != "" && r.connection == "Upgrade"
		checkWebsocketVersionHeader := r.secWebsocketVersion != "" && r.secWebsocketVersion == "13"
		checkWebsocketKeyHeader := r.secWebsocketKey != ""

		if !checkReqMethod || !checkUpgradeHeader || !checkConnectionHeader || !checkWebsocketKeyHeader || !checkWebsocketVersionHeader {
			resp.statusCode = http.StatusBadRequest
			resp.statusText = http.StatusText(resp.statusCode)
			if _, err := conn.Write(resp.Encode()); err != nil {
				fmt.Println(err)
			}
			return
		}

		websocketKey := r.secWebsocketKey
		websocketAcceptValue := CreateWebsocketAcceptValue(websocketKey)

		resp.secWebsocketAccept = websocketAcceptValue
		resp.upgrade = "websocket"
		resp.connection = "Upgrade"
		resp.statusCode = http.StatusSwitchingProtocols
		resp.statusText = http.StatusText(resp.statusCode)
		resp.proto = "HTTP/1.0"

		fmt.Println(resp)

		if _, err := conn.Write(resp.Encode()); err != nil {
			fmt.Println(err)
			return
		}

	*/

	websocketKey := r.secWebsocketKey
	websocketAcceptValue := CreateWebsocketAcceptValue(websocketKey)

	h := &http.Header{}
	h.Add("Upgrade", "websocket")
	h.Add("Connection", "Upgrade")
	h.Add("Sec-WebSocket-Accept", websocketAcceptValue)

	resp := &http.Response{
		StatusCode: http.StatusSwitchingProtocols,
		Status:     http.StatusText(http.StatusSwitchingProtocols),
		Header:     *h,
	}

	if err := resp.Write(conn); err != nil {
		fmt.Println(err)
		return
	}

	ws.conn = conn

	return
}

func CreateWebsocketAcceptValue(secWebsocketKey string) string {
	// Globally Unique Identifier
	GUID := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	keyAndGUIDConcat := fmt.Sprintf("%s%s", secWebsocketKey, GUID)

	secWebsocketKeySHA1 := sha1.Sum([]byte(keyAndGUIDConcat))

	secWebsocketKeyBase64 := base64.StdEncoding.EncodeToString(secWebsocketKeySHA1[:])

	return secWebsocketKeyBase64
}

func parseHandshakeRequest(buffer []byte) *HandshakeRequest {

	req := &HandshakeRequest{}

	bufferString := string(buffer)

	dataString := strings.Split(bufferString, "\n")

	for _, headerData := range dataString {

		if strings.Contains(headerData, " HTTP/") {
			req.method, _, _ = strings.Cut(headerData, " ")
		} else {
			switch header, value, _ := strings.Cut(headerData, ": "); strings.ToLower(header) {
			case "host":
				req.host = strings.TrimSpace(value)
			case "connection":
				req.connection = strings.TrimSpace(value)
			case "sec-websocket-key":
				req.secWebsocketKey = strings.TrimSpace(value)
			case "sec-websocket-version":
				req.secWebsocketVersion = strings.TrimSpace(value)
			case "upgrade":
				req.upgrade = strings.TrimSpace(value)
			}
		}
	}

	return req
}

/*
func (r *HandshakeResponse) Encode() (data []byte) {

	if r.statusCode != http.StatusSwitchingProtocols {
		return []byte(fmt.Sprintf("%s %d %s\n", r.proto, r.statusCode, r.statusText))
	}

	return []byte(fmt.Sprintf("%s %d %s\nConnection: %s\nSec-Websocket-Accept: %s\nUpgrade: %s\n",
		r.proto, r.statusCode, r.statusText, r.connection, r.secWebsocketAccept, r.upgrade))
}

func (r *HandshakeResponse) Decode(data []byte) {
	dataString := string(data)

	r.proto = strings.Split(strings.Split(dataString, "\n")[0], " ")[0]
	r.statusCode, _ = strconv.Atoi(strings.Split(strings.Split(dataString, "\n")[0], " ")[1])
	r.statusText = strings.Split(strings.Split(dataString, "\n")[0], " ")[2] + " " + strings.Split(strings.Split(dataString, "\n")[0], " ")[3]
	r.connection = strings.Split(strings.Split(dataString, "\n")[1], " ")[1]
	r.secWebsocketAccept = strings.Split(strings.Split(dataString, "\n")[2], " ")[1]
	r.upgrade = strings.Split(strings.Split(dataString, "\n")[3], " ")[1]

}
*/
