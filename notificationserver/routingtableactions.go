package notificationserver

import (
	"math/rand"
	"sync"

	"github.com/kubescape/gateway/notificationserver/websocketactions"

	strutils "github.com/armosec/utils-go/str"
	logger "github.com/dwertent/go-logger"
	"github.com/dwertent/go-logger/helpers"

	"github.com/gorilla/websocket"
)

// Connections manages the open websocket connections.
// It acts as a routing table that routes requests to matching connections by
// the attributes provided in requests
type Connections struct {
	connections []*websocketactions.Connection
	mutex       *sync.RWMutex
}

// NewConnectionsObj creates a new Connections object
func NewConnectionsObj() *Connections {
	return &Connections{
		mutex: &sync.RWMutex{},
	}
}

// Append appends a given connection with provided attributes to the current connections
func (cs *Connections) Append(attributes map[string]string, conn *websocket.Conn) (*websocketactions.Connection, int) {
	id := rand.Int()
	connection := websocketactions.NewConnection(conn, id, attributes)
	cs.mutex.Lock()
	cs.connections = append(cs.connections, connection)
	cs.mutex.Unlock()
	return connection, id
}

// Remove removes a connection with given attributes from the routing table
func (cs *Connections) Remove(attributes map[string]string) {
	cs.mutex.Lock()
	slcLen := len(cs.connections)
	for i := 0; i < slcLen; i++ {
		if cs.connections[i].AttributesContained(attributes) {
			logger.L().Info("removing connection from list", helpers.Int("index", i), helpers.String("attributes", strutils.ObjectToString(cs.connections[i].GetAttributes())), helpers.Int("id", cs.connections[i].ID), helpers.Int("list len", len(cs.connections)-1))
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

// RemoveID removes a connection with a given ID from the routing table
func (cs *Connections) RemoveID(id int) {
	cs.mutex.Lock()
	slcLen := len(cs.connections)
	for i := 0; i < slcLen; i++ {
		if cs.connections[i].ID == id {
			logger.L().Info("removing connection from list", helpers.Int("index", i), helpers.String("attributes", strutils.ObjectToString(cs.connections[i].GetAttributes())), helpers.Int("id", cs.connections[i].ID), helpers.Int("list len", len(cs.connections)-1))
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

// Get retrieves a connection with given attributes from the routing table
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

// Len returns the number of the currently managed connections
func (cs *Connections) Len() int {
	cs.mutex.RLocker().Lock()
	l := len(cs.connections)
	cs.mutex.RLocker().Unlock()
	return l
}

// CloseConnections closes all connections that have a set of provided attributes
func (cs *Connections) CloseConnections(wa websocketactions.IWebsocketActions, attributes map[string]string) {
	conns := cs.Get(attributes)
	for i := range conns {
		wa.Close(conns[i])
	}
}
