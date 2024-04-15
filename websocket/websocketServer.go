package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
)

type WebSocketServer struct {}

func (ws *WebSocketServer) handshake(w http.ResponseWriter, r *http.Request) {

	checkReqMethod := r.Method == "GET"
	checkUpgradeHeader := len(r.Header[http.CanonicalHeaderKey("Upgrade")]) == 1 && r.Header[http.CanonicalHeaderKey("Upgrade")][0] == "websocket"
	checkConnectionHeader := len(r.Header[http.CanonicalHeaderKey("Connection")]) == 1 && r.Header[http.CanonicalHeaderKey("Connection")][0] == "Upgrade"
	checkWebsocketVersionHeader := len(r.Header[http.CanonicalHeaderKey("Sec-WebSocket-Version")]) == 1 && r.Header[http.CanonicalHeaderKey("Sec-WebSocket-Version")][0] == "13"
	checkWebsocketKeyHeader := len(r.Header[http.CanonicalHeaderKey("Sec-WebSocket-Key")]) == 1

	if !checkReqMethod || !checkUpgradeHeader || !checkConnectionHeader || !checkWebsocketKeyHeader || !checkWebsocketVersionHeader {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	header := http.CanonicalHeaderKey("Sec-WebSocket-Key")
	websocketKey := r.Header[header][0]

	websocketAcceptValue := CreateWebsocketAcceptValue(websocketKey) 
	
	w.Header().Add("Sec-WebSocket-Accept", websocketAcceptValue)
	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	w.WriteHeader(http.StatusSwitchingProtocols)
	
}

func CreateWebsocketAcceptValue(secWebsocketKey string) string{
	// Globally Unique Identifier
	GUID := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	keyAndGUIDConcat := fmt.Sprintf("%s%s", secWebsocketKey, GUID)

	secWebsocketKeySHA1 := sha1.Sum([]byte(keyAndGUIDConcat))

	secWebsocketKeyBase64 := base64.StdEncoding.EncodeToString(secWebsocketKeySHA1[:])

	return secWebsocketKeyBase64
}
