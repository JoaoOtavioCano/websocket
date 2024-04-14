package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
)

type WebSocketServer struct {}

func (ws *WebSocketServer) handshake(w http.ResponseWriter, r *http.Request) {

	header := http.CanonicalHeaderKey("Sec-WebSocket-Key")
	websocketKey := r.Header[header][0]

	websocketAcceptValue := CreateWebsocketAcceptValue(websocketKey) 
	
	w.Header().Add("Sec-WebSocket-Accept", websocketAcceptValue)
	w.Header().Add("Upgarde", "websocket")
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
