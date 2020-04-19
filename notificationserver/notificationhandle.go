package notificationserver

import (
	"capostman/cautils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"

	"capostman/notificationserver/websocketactions"
)

// NotificationServer -
type NotificationServer struct {
	wa                  websocketactions.IWebsocketActions
	outgoingConnections Connections
	incomingConnections Connections
}

// NewNotificationServer -
func NewNotificationServer() *NotificationServer {
	SetupMasterInfo()
	return &NotificationServer{
		wa:                  &websocketactions.WebsocketActions{},
		outgoingConnections: *NewConnectionsObj(),
		incomingConnections: *NewConnectionsObj(),
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
	log.Printf("recevied query: %s", r.URL.RawQuery)

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
	nh.incomingConnections.Append(notificationAtt, conn)
	defer nh.CleanupIncomeConnection(notificationAtt)

	// ----------------------------------------------------- 3
	// register route in master if master configured
	// create websocket with master
	go nh.ConnectToMaster(notificationAtt)

	// ----------------------------------------------------- 4
	// Websocket read messages
	if err := nh.WebsocketReceiveNotification(conn); err != nil {
		log.Printf("%v, Connection closed", notificationAtt)
	}

}

// ConnectToMaster -
func (nh *NotificationServer) ConnectToMaster(notificationAtt map[string]string) {
	if IsMaster() { // only edge connects to master
		return
	}

	att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)

	// if connected
	if cons := nh.outgoingConnections.Get(att); len(cons) > 0 {
		// update master on new connection?
		return
	}

	masterURL := fmt.Sprintf("%s?", MASTER_HOST)
	amp := ""
	for i, j := range att {
		masterURL += amp
		masterURL += fmt.Sprintf("%s=%s", i, j)
		amp = "&"
	}
	log.Printf("connecting to websocket: %s", masterURL)

	// connect to master
	conn, _, err := nh.wa.DefaultDialer(masterURL, nil)
	if err != nil {
		log.Print(err)
		return
	}
	defer nh.wa.Close(conn)

	// save connection
	nh.outgoingConnections.Append(att, conn)
	defer nh.CleanupOutgoingConnection(att)

	cleanup := make(chan bool)

	// read/write for keeping websocket connection alive
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err := nh.wa.WritePingMessage(conn); err != nil {
				log.Printf("attributes: %v, error: %s", att, err.Error())
				cleanup <- true
			}
		}
	}()
	go func() {
		if err := nh.WebsocketReceiveNotification(conn); err != nil {
			log.Printf("attributes: %v, error: %s", att, err.Error())
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
		log.Printf("%v", err)
		http.Error(w, err.Error(), 400)
		return
	}
	defer r.Body.Close()

	// get notificationID from message
	notificationAtt, err := nh.UnmarshalMessage(readBuffer)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	log.Printf("sending notification to: %v", notificationAtt.Target)

	// set message - add route to message
	if err := nh.SendNotification(notificationAtt.Target, readBuffer); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 400)
	}
}

// SendNotification -
func (nh *NotificationServer) SendNotification(route map[string]string, notification []byte) error {
	connections := nh.incomingConnections.Get(route)
	if len(connections) < 1 {
		return fmt.Errorf("no connections found for attributes: %v", route)
	}
	log.Printf("Posting notification to %v", route)

	for _, conn := range connections {
		err := nh.wa.WriteTextMessage(conn, notification)
		if err != nil {
			log.Printf("%v, connection %p is not alive", route, conn)
			defer nh.CleanupIncomeConnection(route)
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
	log.Printf("target: %v", notificationAtt)

	// TODO: test if route is valid?

	conn, err := nh.wa.ConnectWebsocket(w, r)
	if err != nil {
		return conn, notificationAtt, err
	}

	return conn, notificationAtt, nil

}

// CleanupIncomeConnection -
func (nh *NotificationServer) CleanupIncomeConnection(notificationAtt map[string]string) {
	// remove connection from list
	nh.incomingConnections.Remove(notificationAtt)

	// if edge server (than there is a connection with master server)
	if !IsMaster() {
		att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)
		if len(nh.incomingConnections.Get(att)) < 1 { // there are no more clients connected to edge server with this attributes than disconnect from master
			nh.outgoingConnections.CloseConnections(nh.wa, att)
		}
	}
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
		log.Printf("received msgType: %d. (text=1, close=8, ping=9)", msgType)
		switch msgType {
		case websocket.CloseMessage:
			return fmt.Errorf("websocket recieved CloseMessage")
		case websocket.PingMessage:
			err := nh.wa.WritePongMessage(conn)
			if err != nil {
				return err
			}
		case websocket.TextMessage:
			log.Printf("received message: %s", string(message))

			// get notificationID from message
			n, err := nh.UnmarshalMessage(message)
			if err != nil {
				return err
			}

			// send message
			if err := nh.SendNotification(n.Target, message); err != nil {
				return err
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
