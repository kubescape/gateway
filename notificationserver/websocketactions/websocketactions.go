package websocketactions

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

// IWebsocketActions -
type IWebsocketActions interface {
	ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error)
	WriteBinaryMessage(conn *websocket.Conn, readBuffer []byte) error
	WritePongMessage(conn *websocket.Conn) error
	WritePingMessage(conn *websocket.Conn) error
	ReadMessage(conn *websocket.Conn) (messageType int, p []byte, err error)
	Close(conn *websocket.Conn) error
	DefaultDialer(host string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

// WebsocketActions -
type WebsocketActions struct {
	mutex *sync.Mutex
}

// NewWebsocketActions -
func NewWebsocketActions() *WebsocketActions {
	return &WebsocketActions{
		mutex: &sync.Mutex{},
	}
}

// ConnectWebsocket -
func (wa *WebsocketActions) ConnectWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	return conn, err
}

// WriteBinaryMessage -
func (wa *WebsocketActions) WriteBinaryMessage(conn *websocket.Conn, readBuffer []byte) error {
	wa.mutex.Lock()
	err := conn.WriteMessage(websocket.BinaryMessage, readBuffer)
	wa.mutex.Unlock()
	return err
}

// WritePongMessage -
func (wa *WebsocketActions) WritePongMessage(conn *websocket.Conn) error {
	wa.mutex.Lock()
	err := conn.WriteMessage(websocket.PongMessage, []byte{})
	wa.mutex.Unlock()
	return err
}

// WritePingMessage -
func (wa *WebsocketActions) WritePingMessage(conn *websocket.Conn) error {
	wa.mutex.Lock()
	err := conn.WriteMessage(websocket.PingMessage, []byte{})
	wa.mutex.Unlock()
	return err
}

// ReadMessage -
func (wa *WebsocketActions) ReadMessage(conn *websocket.Conn) (int, []byte, error) {
	wa.mutex.Lock()
	messageType, p, err := conn.ReadMessage()
	wa.mutex.Unlock()
	return messageType, p, err
}

// Close -
func (wa *WebsocketActions) Close(conn *websocket.Conn) error {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorf("recover while closing connection, reason: %v", err)
		}
	}()
	wa.mutex.Lock()
	err := conn.Close()
	wa.mutex.Unlock()
	return err
}

// DefaultDialer -
func (wa *WebsocketActions) DefaultDialer(host string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	i := 0
	for {
		conn, res, err := websocket.DefaultDialer.Dial(host, nil)
		if err != nil {
			err = fmt.Errorf("failed dialing to: '%s', reason: '%s'", host, err.Error())
		}
		if err == nil || i == 2 {
			return conn, res, err
		}
		i++
		glog.Warningf("attempt: %d, error message: %s, waiting 5 seconds before retrying", i, err.Error())
		time.Sleep(time.Second * 5)
	}

}
