package notificationserver

import (
	"sync"

	"github.com/gorilla/websocket"

	"capostman/notificationserver/websocketactions"
)

// Connection -
type Connection struct {
	conn       *websocket.Conn
	attributes map[string]string
}

// Connections -
type Connections struct {
	connections []*Connection
	// attributes   map[string][]*Connection
	mutex *sync.RWMutex
}

// NewConnectionsObj -
func NewConnectionsObj() *Connections {
	return &Connections{
		mutex: &sync.RWMutex{},
	}
}

// Append -
func (cs *Connections) Append(attributes map[string]string, conn *websocket.Conn) {
	cs.mutex.Lock()
	cs.connections = append(cs.connections, &Connection{
		conn:       conn,
		attributes: attributes,
	})
	cs.mutex.Unlock()

}

// Remove from routing table
func (cs *Connections) Remove(attributes map[string]string) {

	cs.mutex.Lock()
	for i := range cs.connections {
		if cs.connections[i].AttributesContained(attributes) {
			cs.connections[i] = cs.connections[len(cs.connections)-1]
			cs.connections = cs.connections[:len(cs.connections)-1]
		}
	}
	cs.mutex.Unlock()
}

// Get from routing table
func (cs *Connections) Get(attributes map[string]string) []*websocket.Conn {
	conns := []*websocket.Conn{}
	cs.mutex.RLocker().Lock()
	for i := range cs.connections {
		if cs.connections[i].AttributesContained(attributes) {
			conns = append(conns, cs.connections[i].conn)
		}
	}
	cs.mutex.RLocker().Unlock()

	return conns
}

// AttributesContained -
func (c *Connection) AttributesContained(attributes map[string]string) bool {
	for i, j := range c.attributes {
		if v, k := attributes[i]; k {
			if v != j {
				return false
			}
		}
	}
	return true
}

// CloseConnections close all connections of set of attributes
func (cs *Connections) CloseConnections(wa websocketactions.IWebsocketActions, attributes map[string]string) {
	conns := cs.Get(attributes)
	for i := range conns {
		defer func() {
			if err := recover(); err != nil {
				cs.mutex.Unlock()
			}
		}()
		cs.mutex.Lock()
		wa.Close(conns[i])
		cs.mutex.Unlock()
	}
}
