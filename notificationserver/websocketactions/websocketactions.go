package websocketactions

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

// IWebsocketActions -
type IWebsocketActions interface {
	ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error)
	WriteTextMessage(conn *websocket.Conn, readBuffer []byte) error
	WritePongMessage(conn *websocket.Conn) error
	ReadMessage(conn *websocket.Conn) (messageType int, p []byte, err error)
}

// WebsocketActions -
type WebsocketActions struct {
}

// ConnectWebsocket -
func (wa *WebsocketActions) ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%v", err)
		return conn, err
	}
	return conn, nil

}

// WriteTextMessage -
func (wa *WebsocketActions) WriteTextMessage(conn *websocket.Conn, readBuffer []byte) error {
	return conn.WriteMessage(websocket.TextMessage, readBuffer)
}

// WritePongMessage -
func (wa *WebsocketActions) WritePongMessage(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.PongMessage, []byte("pong"))
}

// ReadMessage -
func (wa *WebsocketActions) ReadMessage(conn *websocket.Conn) (messageType int, p []byte, err error) {
	return conn.ReadMessage()
}
