package websocketactions

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WebsocketActionsMock -
type WebsocketActionsMock struct {
}

// ConnectWebsocket -
func (wam *WebsocketActionsMock) ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return &websocket.Conn{}, nil
}

// WriteTextMessage -
func (wam *WebsocketActionsMock) WriteTextMessage(conn *websocket.Conn, readBuffer []byte) error {
	return nil
}

// WritePongMessage -
func (wam *WebsocketActionsMock) WritePongMessage(conn *websocket.Conn) error {
	return nil
}

// ReadMessage -
func (wam *WebsocketActionsMock) ReadMessage(conn *websocket.Conn) (messageType int, p []byte, err error) {
	return 1, nil, nil
}
