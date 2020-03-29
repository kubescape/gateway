package notificationserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"

	"capostman/notificationserver/websocketactions"
)

// NotificationHandle -
type NotificationHandle struct {
	wa                  websocketactions.IWebsocketActions
	outgoingConnections Connections
	incomingConnections Connections
}

var (
	// MASTER_ATTRIBUTES attributes master is expecting
	MASTER_ATTRIBUTES []string
	// MASTER_HOST -
	MASTER_HOST string
)

func SetupMasterInfo() {
	att, k1 := os.LookupEnv("MASTER_ATTRIBUTES")
	host, k0 := os.LookupEnv("MASTER_HOST")
	if !k0 || !k1 {
		return
	}
	MASTER_HOST = host
	MASTER_ATTRIBUTES = strings.Split(att, ";")
}

// AppendServer - create websocket with server
func (nh *NotificationHandle) AppendServer(w http.ResponseWriter, r *http.Request) {
	// ----------------------------------------------------- 1
	// receive websocket connetction from client

	if r.Method != "GET" {
		fmt.Printf("Method not allowed")
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, notificationAtt, err := nh.AcceptWebsocketConnection(w, r)
	if err != nil {

	}
	defer conn.Close()
	defer r.Body.Close()

	// ----------------------------------------------------- 2
	// append new route
	nh.incomingConnections.Append(notificationAtt, conn)

	// ----------------------------------------------------- 3
	// register route in master if master configured
	// create websocket with master

	// ----------------------------------------------------- 4
	// Websocket ping pong
	if err := nh.MaintainWebsocket(conn); err != nil {
		log.Printf("%v, Connection closed", notificationAtt)
		defer nh.CleanupConnection(notificationAtt)
		// unregister master?
	}
}

// ConnectToMaster -
func (nh *NotificationHandle) ConnectToMaster(notificationAtt map[string]string) {
	if MASTER_HOST == "" || len(MASTER_ATTRIBUTES) < 1 {
		return
	}

	att := make(map[string]string)
	for i := range MASTER_ATTRIBUTES {
		if v, ok := notificationAtt[MASTER_ATTRIBUTES[i]]; ok {
			att[MASTER_ATTRIBUTES[i]] = v
		}
	}

	// if connected
	if cons := nh.outgoingConnections.Get(att); len(cons) > 0 {
		// update master on new connection?
		return
	}

	// connect to master
	conn, _, err := websocket.DefaultDialer.Dial(MASTER_HOST, nil)
	if err != nil {

	}

	// attach connection
	nh.outgoingConnections.Append(att, conn)
}

// RestAPIReceiveNotification -
func (nh *NotificationHandle) RestAPIReceiveNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", 405)
		return
	}

	defer r.Body.Close()
	readBuffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	notificationAtt, err := nh.ParseURLPath(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// set message - add route to message
	if err := nh.SendNotification(notificationAtt, readBuffer); err != nil {
		http.Error(w, err.Error(), 400)
	}
}

// SendNotification -
func (nh *NotificationHandle) SendNotification(route map[string]string, notification []byte) error {
	connections := nh.incomingConnections.Get(route)
	if len(connections) < 1 {
		return fmt.Errorf("%v, notificationID not found", route)
	}
	log.Printf("%v, Posting notification", route)

	for _, conn := range connections {
		s := string(notification)
		log.Printf("%v:\n%v", route, s)
		err := nh.wa.WriteTextMessage(conn, notification)
		if err != nil {
			// Remove connection
			log.Printf("%v, connection %p is not alive", route, conn)
			defer nh.CleanupConnection(route)
		}
	}
	return nil
}

// HandleEdgeServerNotification -
func (nh *NotificationHandle) HandleEdgeServerNotification() {
	// receive message from client

	// handle message
	// - ping
	// - unregister
	// - status
}

// RemoveServer -
func (nh *NotificationHandle) RemoveServer() {
	// unregister from master

}

// AcceptWebsocketConnection -
func (nh *NotificationHandle) AcceptWebsocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, map[string]string, error) {

	notificationAtt, err := nh.ParseURLPath(r)
	if err != nil {
		return nil, notificationAtt, err
	}

	// TODO: test the route is valid

	conn, err := nh.wa.ConnectWebsocket(w, r)
	if err != nil {
		return conn, notificationAtt, err
	}

	return conn, notificationAtt, nil

}

// CleanupConnection -
func (nh *NotificationHandle) CleanupConnection(notificationAtt map[string]string) {
	nh.incomingConnections.Remove(notificationAtt)
	// unregister from master?
}

// MaintainWebsocket maintain websocket connection // RestAPIReceiveNotification -
func (nh *NotificationHandle) MaintainWebsocket(conn *websocket.Conn) error {
	// Websocket ping pong
	for {
		msgType, _, err := nh.wa.ReadMessage(conn)
		if err != nil {
			return err
		}

		switch msgType {
		case websocket.CloseMessage:
			return err
		case websocket.PingMessage:
			err = nh.wa.WritePongMessage(conn)
			if err != nil {
				return err
			}
		case websocket.TextMessage:
			// get notificationID from message
			// notificationID := ""

			// send message
			// sendNotification(notificationID, notification)
		}
	}
}

// ParseURLPath -
func (nh *NotificationHandle) ParseURLPath(r *http.Request) (map[string]string, error) {
	return map[string]string{}, nil

	// urlPath := strings.Split(r.URL.Path, "/")[2:]
	// if len(urlPath) < 1 {
	// 	return urlPath, fmt.Errorf("")
	// }
	// return urlPath, nil
}
