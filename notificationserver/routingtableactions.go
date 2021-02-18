package notificationserver

import (
	"math/rand"
	"notification-server/notificationserver/websocketactions"
	"sync"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

// Connections -
type Connections struct {
	connections []*websocketactions.Connection
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
func (cs *Connections) Append(attributes map[string]string, conn *websocket.Conn) (*websocketactions.Connection, int) {
	id := rand.Int()
	connection := websocketactions.NewConnection(conn, id, attributes)
	cs.mutex.Lock()
	cs.connections = append(cs.connections, connection)
	cs.mutex.Unlock()
	return connection, id
}

// Remove from routing table
func (cs *Connections) Remove(attributes map[string]string) {
	cs.mutex.Lock()
	slcLen := len(cs.connections)
	for i := 0; i < slcLen; i++ {
		if cs.connections[i].AttributesContained(attributes) {
			glog.Infof("Removing connection from list. index: %d. attributes: %v, id: %d, list len: %d", i, cs.connections[i].GetAttributes(), cs.connections[i].ID, len(cs.connections)-1)
			if slcLen == 1 { //i is the only element in the slice so we need to remove this entry from the map
				cs.connections = []*websocketactions.Connection{}
			} else if i == slcLen-1 { // i is the last element in the slice so i+1 is out of range
				cs.connections = cs.connections[:i]
			} else {
				cs.connections = append(cs.connections[:i], cs.connections[i+1:]...)
			}
			slcLen--
			i--
		}
	}
	cs.mutex.Unlock()
}

// RemoveID by id from routing table
func (cs *Connections) RemoveID(id int) {
	cs.mutex.Lock()
	slcLen := len(cs.connections)
	for i := 0; i < slcLen; i++ {
		if cs.connections[i].ID == id {
			glog.Infof("Removing connection from list. index: %d. attributes: %v, id: %d, list len: %d", i, cs.connections[i].GetAttributes(), cs.connections[i].ID, len(cs.connections)-1)
			if slcLen == 1 { //i is the only element in the slice so we need to remove this entry from the map
				cs.connections = []*websocketactions.Connection{}
			} else if i == slcLen-1 { // i is the last element in the slice so i+1 is out of range
				cs.connections = cs.connections[:i]
			} else {
				cs.connections = append(cs.connections[:i], cs.connections[i+1:]...)
			}
			slcLen--
			i--
		}
	}
	cs.mutex.Unlock()
}

// Get from routing table
func (cs *Connections) Get(attributes map[string]string) []*websocketactions.Connection {
	conns := []*websocketactions.Connection{}
	cs.mutex.RLocker().Lock()
	for i := range cs.connections {
		if cs.connections[i].AttributesContained(attributes) {
			conns = append(conns, cs.connections[i])
		}
	}
	cs.mutex.RLocker().Unlock()
	return conns
}

// Len list length
func (cs *Connections) Len() int {
	cs.mutex.RLocker().Lock()
	l := len(cs.connections)
	cs.mutex.RLocker().Unlock()
	return l
}

// CloseConnections close all connections of set of attributes
func (cs *Connections) CloseConnections(wa websocketactions.IWebsocketActions, attributes map[string]string) {
	conns := cs.Get(attributes)
	for i := range conns {
		wa.Close(conns[i])
	}
}
