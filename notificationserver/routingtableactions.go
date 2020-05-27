package notificationserver

import (
	"log"
	"math/rand"
	"notificationserver/notificationserver/websocketactions"
	"sync"

	"github.com/gorilla/websocket"
)

// Connection -
type Connection struct {
	ID         int
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
func (cs *Connections) Append(attributes map[string]string, conn *websocket.Conn) int {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	id := rand.Int()
	cs.connections = append(cs.connections, &Connection{
		ID:         id,
		conn:       conn,
		attributes: attributes,
	})
	return id
}

// Remove from routing table
func (cs *Connections) Remove(attributes map[string]string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	slcLen := len(cs.connections)
	for i := 0; i < slcLen; i++ {
		if cs.connections[i].AttributesContained(attributes) {
			log.Printf("Removing connection from incoming list: %d. attributes: %v", i, attributes)
			if slcLen < 2 { //i is the only element in the slice so we need to remove this entry from the map
				cs.connections = make([]*Connection, 0, 10)
			} else if i == slcLen-1 { // i is the last element in the slice so i+1 is out of range
				cs.connections = cs.connections[:i]
			} else {
				cs.connections = append(cs.connections[:i], cs.connections[i+1:]...)
			}
			slcLen--
			i--
		}
	}
}

// RemoveID by id from routing table
func (cs *Connections) RemoveID(id int) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	slcLen := len(cs.connections)
	for i := 0; i < slcLen; i++ {
		if cs.connections[i].ID == id {
			log.Printf("Removing connection from incoming list. index: %d. attributes: %v, id: %d", i, cs.connections[i].attributes, cs.connections[i].ID)
			if slcLen < 2 { //i is the only element in the slice so we need to remove this entry from the map
				cs.connections = make([]*Connection, 0, 10)
			} else if i == slcLen-1 { // i is the last element in the slice so i+1 is out of range
				cs.connections = cs.connections[:i]
			} else {
				cs.connections = append(cs.connections[:i], cs.connections[i+1:]...)
			}
			slcLen--
			i--
		}
	}
}

// Get from routing table
func (cs *Connections) Get(attributes map[string]string) []*Connection {
	conns := []*Connection{}
	cs.mutex.RLocker().Lock()
	defer cs.mutex.RLocker().Unlock()
	for i := range cs.connections {
		if cs.connections[i].AttributesContained(attributes) {
			conns = append(conns, cs.connections[i])
		}
	}

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
		wa.Close(conns[i].conn)
		cs.mutex.Unlock()
	}
}
