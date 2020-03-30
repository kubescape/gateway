package notificationserver

import (
	"capostman/cautils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"capostman/notificationserver/websocketactions"
)

// NotificationHandle -
type NotificationHandle struct {
	wa                  websocketactions.IWebsocketActions
	outgoingConnections Connections
	incomingConnections Connections
}

// Notification passed between servers
type Notification struct {
	Target       map[string]string `json:"target"`
	Notification interface{}       `json:"notification"`
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
	if MASTER_ATTRIBUTES[len(MASTER_ATTRIBUTES)-1] == "" {
		cautils.RemoveIndexFromStringSlice(&MASTER_ATTRIBUTES, len(MASTER_ATTRIBUTES)-1)
	}
}

// IsMaster is server master or edge
func IsMaster() bool {
	return (len(MASTER_ATTRIBUTES) > 0 && MASTER_HOST != "")
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
	go func() {
		if err := nh.ConnectToMaster(notificationAtt); err != nil {
			att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)
			nh.CleanupOutgoingConnection(att)
		}
	}()

	// ----------------------------------------------------- 4
	// Websocket read messages
	if err := nh.WebsocketReceiveNotification(conn); err != nil {
		log.Printf("%v, Connection closed", notificationAtt)
		defer nh.CleanupIncomeConnection(notificationAtt)
	}
}

// ConnectToMaster -
func (nh *NotificationHandle) ConnectToMaster(notificationAtt map[string]string) error {
	if IsMaster() { // only edge connects to master
		return nil
	}

	att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)

	// if connected
	if cons := nh.outgoingConnections.Get(att); len(cons) > 0 {
		// update master on new connection?
		return nil
	}

	// connect to master
	conn, _, err := nh.wa.DefaultDialer(MASTER_HOST, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// save connection
	nh.outgoingConnections.Append(att, conn)

	// read/write for keeping websocket connection alive
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err = nh.wa.WritePingMessage(conn); err != nil {
				nh.CleanupOutgoingConnection(att)
			}
		}
	}()
	return nh.WebsocketReceiveNotification(conn)
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
func (nh *NotificationHandle) SendNotification(route map[string]string, notification interface{}) error {
	connections := nh.incomingConnections.Get(route)
	if len(connections) < 1 {
		return fmt.Errorf("%v, notificationID not found", route)
	}
	log.Printf("%v, Posting notification", route)

	var k bool
	var s string
	var notif []byte

	if notif, k = notification.([]byte); k {
		// s = string(notif)
	} else if s, k = notification.(string); k {
		notif = []byte(s)
	} else {
		return fmt.Errorf("unknown type notification. received: %v", notification)
	}
	for _, conn := range connections {
		err := nh.wa.WriteTextMessage(conn, notif)
		if err != nil {
			// Remove connection
			log.Printf("%v, connection %p is not alive", route, conn)
			defer nh.CleanupIncomeConnection(route)
		}
	}
	return nil
}

// HandleServerNotification -
func (nh *NotificationHandle) HandleServerNotification() {
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

	// TODO: test the route is valid?

	conn, err := nh.wa.ConnectWebsocket(w, r)
	if err != nil {
		return conn, notificationAtt, err
	}

	return conn, notificationAtt, nil

}

// CleanupIncomeConnection -
func (nh *NotificationHandle) CleanupIncomeConnection(notificationAtt map[string]string) {
	// remove connection from list
	nh.incomingConnections.Remove(notificationAtt)

	// if edge server (than there is a connection with master server)
	if !IsMaster() {
		att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)
		if len(nh.incomingConnections.Get(att)) < 1 { // there are no more clients connected to edge server with this attributes than disconnect from master
			nh.outgoingConnections.CloseConnections(att)
		}
	}
}

// CleanupOutgoingConnection -
func (nh *NotificationHandle) CleanupOutgoingConnection(notificationAtt map[string]string) {
	// remove outgoing connection from list
	nh.outgoingConnections.Remove(notificationAtt)

	// close all incoming connections relaited to this attributes
	nh.incomingConnections.CloseConnections(notificationAtt)
}

// WebsocketReceiveNotification maintain websocket connection // RestAPIReceiveNotification -
func (nh *NotificationHandle) WebsocketReceiveNotification(conn *websocket.Conn) error {
	// Websocket ping pong
	for {
		msgType, message, err := nh.wa.ReadMessage(conn)
		if err != nil {
			return err
		}

		switch msgType {
		case websocket.CloseMessage:
			return fmt.Errorf("websocket recieved CloseMessage")
		case websocket.PingMessage:
			err = nh.wa.WritePongMessage(conn)
			if err != nil {
				return err
			}
		case websocket.TextMessage:
			// get notificationID from message
			n := Notification{}
			if err := json.Unmarshal(message, &n); err != nil {
				return err
			}
			// send message
			if err := nh.SendNotification(n.Target, n.Notification); err != nil {
				return err
			}
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
