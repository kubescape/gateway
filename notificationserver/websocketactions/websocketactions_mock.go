package websocketactions

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	READ_MESSAGE_TYPE_MOCK = websocket.CloseMessage
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

// WritePingMessage -
func (wam *WebsocketActionsMock) WritePingMessage(conn *websocket.Conn) error {
	return nil
}

// ReadMessage -
func (wam *WebsocketActionsMock) ReadMessage(conn *websocket.Conn) (messageType int, p []byte, err error) {
	return READ_MESSAGE_TYPE_MOCK, nil, nil
}

// DefaultDialer -
func (wam *WebsocketActionsMock) DefaultDialer(host string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	return &websocket.Conn{}, nil, nil
}

// Close -
func (wam *WebsocketActionsMock) Close(conn *websocket.Conn) error {
	return nil
}
