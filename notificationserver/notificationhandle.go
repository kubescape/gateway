package notificationserver

import (
	"notificationserver/cautils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"notificationserver/notificationserver/websocketactions"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
)

// NotificationServer -
type NotificationServer struct {
	wa                       websocketactions.IWebsocketActions
	outgoingConnections      Connections
	incomingConnections      Connections
	outgoingConnectionsMutex *sync.Mutex
}

// NewNotificationServer -
func NewNotificationServer() *NotificationServer {
	SetupMasterInfo()
	return &NotificationServer{
		wa:                       &websocketactions.WebsocketActions{},
		outgoingConnections:      *NewConnectionsObj(),
		incomingConnections:      *NewConnectionsObj(),
		outgoingConnectionsMutex: &sync.Mutex{},
	}
}

// Notification passed between servers
type Notification struct {
	Target       map[string]string `json:"target"`
	Notification interface{}       `json:"notification"`
}

// WebsocketNotificationHandler - create websocket with server
func (nh *NotificationServer) WebsocketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	// ----------------------------------------------------- 1
	// receive websocket connetction from client
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed")
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, notificationAtt, err := nh.AcceptWebsocketConnection(w, r)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 400)
		return
	}
	defer nh.wa.Close(conn)

	// ----------------------------------------------------- 2
	// append new route
	id := nh.incomingConnections.Append(notificationAtt, conn)
	defer nh.CleanupIncomeConnection(id)

	log.Printf("accepting websocket connection. url query: %s, id: %d", r.URL.RawQuery, id)

	// ----------------------------------------------------- 3
	// register route in master if master configured
	// create websocket with master
	go nh.ConnectToMaster(notificationAtt)

	// ----------------------------------------------------- 4
	// Websocket read messages
	if err := nh.WebsocketReceiveNotification(conn); err != nil {
		log.Printf("In WebsocketNotificationHandler connection closed. attributes: %v, error: %v", notificationAtt, err)
	}

}

// ConnectToMaster -
func (nh *NotificationServer) ConnectToMaster(notificationAtt map[string]string) {
	if IsMaster() { // only edge connects to master
		return
	}

	att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)

	nh.outgoingConnectionsMutex.Lock()

	// if connected
	if cons := nh.outgoingConnections.Get(att); len(cons) > 0 {
		log.Printf("edge already connected to master, not creating new connection")
		// update master on new connection?
		nh.outgoingConnectionsMutex.Unlock()
		return
	}

	masterURL := fmt.Sprintf("%s?", MASTER_HOST)
	amp := ""
	for i, j := range att {
		masterURL += amp
		masterURL += fmt.Sprintf("%s=%s", i, j)
		amp = "&"
	}
	log.Printf("connecting to master: %s", masterURL)

	// connect to master
	conn, _, err := nh.wa.DefaultDialer(masterURL, nil)
	if err != nil {
		log.Printf("In ConnectToMaster: %v", err)
		nh.outgoingConnectionsMutex.Unlock()
		return
	}
	defer nh.wa.Close(conn)
	nh.outgoingConnections.Append(att, conn)
	nh.outgoingConnectionsMutex.Unlock()

	defer nh.CleanupOutgoingConnection(att)

	cleanup := make(chan bool)

	// read/write for keeping websocket connection alive
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err := nh.wa.WritePingMessage(conn); err != nil {
				log.Printf("In WritePingMessage attributes: %v, error: %s", att, err.Error())
				cleanup <- true
			}
		}
	}()
	go func() {
		if err := nh.WebsocketReceiveNotification(conn); err != nil {
			log.Printf("In ConnectToMaster WebsocketReceiveNotification attributes: %v, error: %s", att, err.Error())
			cleanup <- true
		}
	}()

	<-cleanup
	log.Printf("disconnect from master with connection attributes: %v", att)
}

// RestAPINotificationHandler -
func (nh *NotificationServer) RestAPINotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", 405)
		return
	}

	readBuffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("In RestAPINotificationHandler ReadAll %v", err)
		http.Error(w, err.Error(), 400)
		return
	}
	defer r.Body.Close()

	// get notificationID from message
	notificationAtt, err := nh.UnmarshalMessage(readBuffer)
	if err != nil {
		log.Printf("In RestAPINotificationHandler UnmarshalMessage %v", err)
		http.Error(w, err.Error(), 400)
		return
	}
	log.Printf("REST API received message, attributes: %v", notificationAtt.Target)

	// set message - add route to message
	if err := nh.SendNotification(notificationAtt.Target, readBuffer); err != nil {
		log.Printf("In RestAPINotificationHandler SendNotification %v, target: %v", err, notificationAtt.Target)
		http.Error(w, err.Error(), 400)
	}
}

// SendNotification -
func (nh *NotificationServer) SendNotification(route map[string]string, notification []byte) error {

	connections := nh.incomingConnections.Get(route)
	for _, conn := range connections {
		log.Printf("sending notification to: %v, id: %d", route, conn.ID)
		err := nh.wa.WriteTextMessage(conn.conn, notification)
		if err != nil {
			log.Printf("In SendNotification %v, connection %d is not alive, error: %v", route, conn.ID, err)
			defer nh.CleanupIncomeConnection(conn.ID)
		}
	}
	return nil
}

// AcceptWebsocketConnection -
func (nh *NotificationServer) AcceptWebsocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, map[string]string, error) {

	notificationAtt, err := nh.ParseURLPath(r.URL)
	if err != nil {
		return nil, notificationAtt, err
	}

	conn, err := nh.wa.ConnectWebsocket(w, r)
	if err != nil {
		return conn, notificationAtt, err
	}

	return conn, notificationAtt, nil

}

// CleanupIncomeConnections - cleanup all connections with matching attributes
func (nh *NotificationServer) CleanupIncomeConnections(notificationAtt map[string]string) {
	// remove connection from list
	nh.incomingConnections.Remove(notificationAtt)

	// TODO- find a way to cleanup connections with master
	// // if edge server (than there is a connection with master server)
	// if !IsMaster() {
	// 	att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)
	// 	if len(nh.incomingConnections.Get(att)) < 1 { // there are no more clients connected to edge server with this attributes than disconnect from master
	// 		nh.outgoingConnections.CloseConnections(nh.wa, att)
	// 	}
	// }
}

// CleanupIncomeConnection - cleanup one connection with matching id
func (nh *NotificationServer) CleanupIncomeConnection(id int) {
	// remove connection from list
	nh.incomingConnections.RemoveID(id)

}

// CleanupOutgoingConnection -
func (nh *NotificationServer) CleanupOutgoingConnection(notificationAtt map[string]string) {
	// remove outgoing connection from list
	nh.outgoingConnections.Remove(notificationAtt)

	// close all incoming connections relaited to this attributes
	nh.incomingConnections.CloseConnections(nh.wa, notificationAtt)
}

// WebsocketReceiveNotification maintain websocket connection // RestAPIReceiveNotification -
func (nh *NotificationServer) WebsocketReceiveNotification(conn *websocket.Conn) error {
	// Websocket ping pong
	for {
		msgType, message, err := nh.wa.ReadMessage(conn)
		if err != nil {
			return err
		}
		// log.Printf("In WebsocketReceiveNotification received msgType: %d. (text=1, close=8, ping=9)", msgType)
		switch msgType {
		case websocket.CloseMessage:
			return fmt.Errorf("In WebsocketReceiveNotification websocket recieved CloseMessage")
		case websocket.PingMessage:
			err := nh.wa.WritePongMessage(conn)
			if err != nil {
				return fmt.Errorf("In WebsocketReceiveNotification WritePongMessage error: %v", err)
			}
		case websocket.TextMessage:
			// get notificationID from message
			n, err := nh.UnmarshalMessage(message)
			if err != nil {
				return fmt.Errorf("In WebsocketReceiveNotification UnmarshalMessage error: %v", err)
			}

			// send message
			if err := nh.SendNotification(n.Target, message); err != nil {
				return fmt.Errorf("In WebsocketReceiveNotification SendNotification error: %v", err)
			}
		}
	}
}

// ParseURLPath -
func (nh *NotificationServer) ParseURLPath(u *url.URL) (map[string]string, error) {
	// keeping backward compatibility (capostman)
	urlPath := strings.Split(u.Path, "/")
	if len(urlPath) == 4 && urlPath[3] != "" { // capostamn
		return map[string]string{urlPath[3]: ""}, nil
	}

	att := make(map[string]string)
	q := u.Query()
	for k, v := range q {
		if k != "" && len(v) > 0 {
			att[k] = v[0]
		}
	}
	if len(att) < 1 {
		return att, fmt.Errorf("no attributes received")
	}
	return att, nil
}

// UnmarshalMessage -
func (nh *NotificationServer) UnmarshalMessage(message []byte) (*Notification, error) {
	n := &Notification{}
	var err error
	if err = json.Unmarshal(message, n); err == nil {
		return n, nil
	}
	if err = bson.Unmarshal(message, n); err == nil {
		return n, nil
	}
	return n, err
}
