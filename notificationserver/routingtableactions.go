package notificationserver

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Connection -
type Connection struct {
	conn       *websocket.Conn
	attributes map[string]string
}

// Connections -
type Connections struct {
	routingTable []*Connection
	// attributes   map[string][]*Connection
	mutex sync.Mutex
}

// Append -
func (cs *Connections) Append(attributes map[string]string, conn *websocket.Conn) {
	cs.mutex.Lock()
	cs.routingTable = append(cs.routingTable, &Connection{
		conn:       conn,
		attributes: attributes,
	})
	cs.mutex.Unlock()

}

// Remove from routing table
func (cs *Connections) Remove(attributes map[string]string) {

	for i := range cs.routingTable {
		cs.mutex.Lock()
		if cs.routingTable[i].AttributesContained(attributes) {
			cs.routingTable[i] = cs.routingTable[len(cs.routingTable)-1]
			cs.routingTable = cs.routingTable[:len(cs.routingTable)-1]
		}
		cs.mutex.Unlock()
	}
}

// Get from routing table
func (cs *Connections) Get(attributes map[string]string) []*websocket.Conn {
	conns := []*websocket.Conn{}
	// cs.mutex.Lock()
	for i := range cs.routingTable {
		if cs.routingTable[i].AttributesContained(attributes) {
			conns = append(conns, cs.routingTable[i].conn)
		}
	}
	// cs.mutex.Unlock()

	return conns
}

// AttributesContained -
func (c *Connection) AttributesContained(attributes map[string]string) bool {
	for i, j := range attributes {
		if v, k := c.attributes[i]; !k || v != j {
			return false
		}
	}
	return true
}

// CloseConnections close all connections of set of attributes
func (cs *Connections) CloseConnections(attributes map[string]string) {
	conns := cs.Get(attributes)
	for i := range conns {
		defer func() {
			if err := recover(); err != nil {
				cs.mutex.Unlock()
			}
		}()
		cs.mutex.Lock()
		conns[i].Close()
		cs.mutex.Unlock()
	}
}
