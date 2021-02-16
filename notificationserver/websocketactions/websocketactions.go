package websocketactions

import (
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
	wa.mutex.Lock()
	defer wa.mutex.Unlock()
	return upgrader.Upgrade(w, r, nil)
}

// WriteBinaryMessage -
func (wa *WebsocketActions) WriteBinaryMessage(conn *websocket.Conn, readBuffer []byte) error {
	wa.mutex.Lock()
	defer wa.mutex.Unlock()
	return conn.WriteMessage(websocket.BinaryMessage, readBuffer)
}

// WritePongMessage -
func (wa *WebsocketActions) WritePongMessage(conn *websocket.Conn) error {
	wa.mutex.Lock()
	defer wa.mutex.Unlock()
	return conn.WriteMessage(websocket.PongMessage, []byte{})
}

// WritePingMessage -
func (wa *WebsocketActions) WritePingMessage(conn *websocket.Conn) error {
	wa.mutex.Lock()
	defer wa.mutex.Unlock()
	return conn.WriteMessage(websocket.PingMessage, []byte{})
}

// ReadMessage -
func (wa *WebsocketActions) ReadMessage(conn *websocket.Conn) (messageType int, p []byte, err error) {
	wa.mutex.Lock()
	defer wa.mutex.Unlock()
	return conn.ReadMessage()
}

// Close -
func (wa *WebsocketActions) Close(conn *websocket.Conn) error {
	return conn.Close()
}

// DefaultDialer -
func (wa *WebsocketActions) DefaultDialer(host string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	i := 0
	for {
		wa.mutex.Lock()
		conn, res, err := websocket.DefaultDialer.Dial(host, nil)
		if err == nil || i == 2 {
			wa.mutex.Unlock()
			return conn, res, err
		}
		wa.mutex.Unlock()
		i++
		glog.Infof("attempt: %d, error message: %s", i, err.Error())
		time.Sleep(time.Second * 5)
	}

}
