package websocketactions

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WebsocketActionsMock -
type WebsocketActionsMock struct {
}

// ReadMessageTypeMock mock message
var ReadMessageTypeMock = websocket.CloseMessage

// ConnectWebsocket -
func (wam *WebsocketActionsMock) ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return &websocket.Conn{}, nil
}

// WriteBinaryMessage -
func (wam *WebsocketActionsMock) WriteBinaryMessage(conn *Connection, readBuffer []byte) error {
	return nil
}

// WritePreparedMessage -
func (wam *WebsocketActionsMock) WritePreparedMessage(conn *Connection, preparedMessage *websocket.PreparedMessage) error {
	return nil
}

// WritePongMessage -
func (wam *WebsocketActionsMock) WritePongMessage(conn *Connection) error {
	return nil
}

// WritePingMessage -
func (wam *WebsocketActionsMock) WritePingMessage(conn *Connection) error {
	return nil
}

// ReadMessage -
func (wam *WebsocketActionsMock) ReadMessage(conn *Connection) (int, []byte, error) {
	return ReadMessageTypeMock, nil, nil
}

// DefaultDialer -
func (wam *WebsocketActionsMock) DefaultDialer(host string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	return &websocket.Conn{}, nil, nil
}

// Close -
func (wam *WebsocketActionsMock) Close(conn *Connection) error {
	return nil
}
