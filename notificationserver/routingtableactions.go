package notificationserver

import "github.com/gorilla/websocket"

// Connection -
type Connection struct {
	conn       *websocket.Conn
	attributes map[string]string
}

// Connections -
type Connections struct {
	routingTable []*Connection
	attributes   map[string][]*Connection
}

// Append -
func (cs *Connections) Append(attributes map[string]string, conn *websocket.Conn) {
	cs.routingTable = append(cs.routingTable, &Connection{
		conn:       conn,
		attributes: attributes,
	})
}

// Remove from routing table
func (cs *Connections) Remove(route map[string]string) {
	// for index, element := range notificationMap[notificationID] {
	// 	if element == conn {
	// 		fmt.Printf("%s, Removing notification", notificationID)
	// 		notificationMap[notificationID] = remove(notificationMap[notificationID], index)
	// 		return
	// 	}
	// }
	/*
		func remove(s []*websocket.Conn, i int) []*websocket.Conn {
			s[i] = s[len(s)-1]
			return s[:len(s)-1]
		}
	*/
}

// Get from routing table
func (cs *Connections) Get(attributes map[string]string) []*websocket.Conn {
	conns := []*websocket.Conn{}
	for i := range cs.routingTable {
		if cs.routingTable[i].AttributesContained(attributes) {
			conns = append(conns, cs.routingTable[i].conn)
		}
	}
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
