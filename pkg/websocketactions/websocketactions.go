package websocketactions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	logger "github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

// IWebsocketActions -
type IWebsocketActions interface {
	ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error)
	WriteBinaryMessage(conn *Connection, readBuffer []byte) error
	WritePongMessage(conn *Connection) error
	WritePingMessage(conn *Connection) error
	WritePreparedMessage(conn *Connection, preparedMessage *websocket.PreparedMessage) error
	ReadMessage(conn *Connection) (int, []byte, error)
	Close(conn *Connection) error
	DefaultDialer(host string, headers http.Header) (*websocket.Conn, *http.Response, error)
}

// WebsocketActions -
type WebsocketActions struct {
}

// NewWebsocketActions -
func NewWebsocketActions() *WebsocketActions {
	return &WebsocketActions{}
}

// ConnectWebsocket -
func (wa *WebsocketActions) ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	return conn, err
}

// WriteBinaryMessage -
func (wa *WebsocketActions) WriteBinaryMessage(conn *Connection, readBuffer []byte) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	err := conn.conn.WriteMessage(websocket.BinaryMessage, readBuffer)
	return err
}

// WritePreparedMessage -
func (wa *WebsocketActions) WritePreparedMessage(conn *Connection, preparedMessage *websocket.PreparedMessage) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	err := conn.conn.WritePreparedMessage(preparedMessage)
	return err
}

// WritePongMessage -
func (wa *WebsocketActions) WritePongMessage(conn *Connection) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	err := conn.conn.WriteMessage(websocket.PongMessage, []byte{})
	return err
}

// WritePingMessage -
func (wa *WebsocketActions) WritePingMessage(conn *Connection) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	err := conn.conn.WriteMessage(websocket.PingMessage, []byte{})
	return err
}

// ReadMessage -
func (wa *WebsocketActions) ReadMessage(conn *Connection) (int, []byte, error) {
	messageType, p, err := conn.conn.ReadMessage()
	return messageType, p, err
}

// Close -
func (wa *WebsocketActions) Close(conn *Connection) error {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	defer func() {
		if err := recover(); err != nil {
			logger.L().Error("recover while closing connection", helpers.Interface("reason", err))
		}
	}()
	err := conn.conn.Close()
	return err
}

// DefaultDialer -
func (wa *WebsocketActions) DefaultDialer(host string, headers http.Header) (*websocket.Conn, *http.Response, error) {
	i := 0
	for {
		conn, res, err := websocket.DefaultDialer.Dial(host, headers)
		if err != nil {
			err = fmt.Errorf("failed dialing to: '%s', reason: '%s'", host, err.Error())
		}
		if err == nil || i == 2 {
			return conn, res, err
		}
		i++
		logger.L().Warning("dialing websocket", helpers.Int("attempt", i), helpers.Error(err))
		time.Sleep(time.Second * 5)
	}

}
