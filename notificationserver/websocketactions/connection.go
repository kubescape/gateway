package websocketactions

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Connection -
type Connection struct {
	mutex      *sync.Mutex
	ID         int
	conn       *websocket.Conn
	attributes map[string]string
}

// NewConnection -
func NewConnection(conn *websocket.Conn, id int, attributes map[string]string) *Connection {
	return &Connection{
		mutex:      &sync.Mutex{},
		ID:         id,
		conn:       conn,
		attributes: attributes,
	}
}

// GetAttributes -
func (c *Connection) GetAttributes() map[string]string {
	return c.attributes
}

// AttributesContained -
func (c *Connection) AttributesContained(attributes map[string]string) bool {
	found := false
	for i, j := range c.attributes {
		if v, k := attributes[i]; k {
			if v != j {
				return false
			}
			found = true
		}
	}
	return found
}
