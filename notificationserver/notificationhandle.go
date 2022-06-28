package notificationserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"notification-server/cautils"
	"os"
	"strings"
	"sync"
	"time"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	"github.com/armosec/utils-k8s-go/armometadata"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"

	"notification-server/notificationserver/websocketactions"
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
	pathToConfig := os.Getenv("CA_CONFIG") // if empty, will load config from default path
	if _, err := armometadata.LoadConfig(pathToConfig, true); err != nil {
		glog.Warning(err.Error())
	}

	SetupMasterInfo()
	return &NotificationServer{
		wa:                       websocketactions.NewWebsocketActions(),
		outgoingConnections:      *NewConnectionsObj(),
		incomingConnections:      *NewConnectionsObj(),
		outgoingConnectionsMutex: &sync.Mutex{},
	}
}

type Notification notifier.Notification

// WebsocketNotificationHandler - create websocket with server
func (nh *NotificationServer) WebsocketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	// ----------------------------------------------------- 1
	// receive websocket connection from client
	if r.Method != http.MethodGet {
		glog.Errorf("Method not allowed")
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, notificationAtt, err := nh.AcceptWebsocketConnection(w, r)
	if err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 400)
		return
	}

	// ----------------------------------------------------- 2
	// append new route
	newConn, id := nh.incomingConnections.Append(notificationAtt, conn)
	glog.Infof("accepting websocket connection. url query: %s, id: %d, number of incoming websockets: %d", r.URL.RawQuery, id, nh.incomingConnections.Len())

	// ----------------------------------------------------- 3
	// register route in master if master configured
	// create websocket with master
	go nh.ConnectToMaster(notificationAtt, 0)

	// ----------------------------------------------------- 4
	// Websocket read messages
	if err := nh.WebsocketReceiveNotification(newConn); err != nil {
		if !strings.Contains(err.Error(), "CloseMessage") {
			// nh.wa.Close(conn)
		}
	}
	nh.CleanupIncomeConnection(id)
	nh.wa.Close(newConn)
}

// ConnectToMaster -
func (nh *NotificationServer) ConnectToMaster(notificationAtt map[string]string, retry int) {
	if IsMaster() { // only edge connects to master
		return
	}

	att := cautils.MergeSliceAndMap(MASTER_ATTRIBUTES, notificationAtt)
	if len(att) == 0 {
		att = notificationAtt
	}
	nh.outgoingConnectionsMutex.Lock() // lock connecting to master to prevent many connections

	// if connected
	if cons := nh.outgoingConnections.Get(att); len(cons) > 0 {
		nh.outgoingConnectionsMutex.Unlock()
		glog.Infof("edge already connected to master, not creating new connection")
		return
	}

	masterURL := fmt.Sprintf("%s?", MASTER_HOST)
	amp := ""
	for i, j := range att {
		masterURL += amp
		masterURL += fmt.Sprintf("%s=%s", i, j)
		amp = "&"
	}
	glog.Infof("connecting to master: %s", masterURL)

	// connect to master
	conn, _, err := nh.wa.DefaultDialer(masterURL, nil)
	if err != nil {
		nh.outgoingConnectionsMutex.Unlock()
		glog.Errorf("in ConnectToMaster: %v", err)
		return
	}
	connObj, _ := nh.outgoingConnections.Append(att, conn)
	nh.outgoingConnectionsMutex.Unlock()

	glog.Infof("successfully contented to master, number of outgoing websockets: %d", nh.outgoingConnections.Len())

	// read/write for keeping websocket connection alive
	go func(pconnObj *websocketactions.Connection) {
		for {
			time.Sleep(10 * time.Second)
			if err := nh.wa.WritePingMessage(pconnObj); err != nil {
				glog.Warningf("in WritePingMessage attributes: %v, reason: %s", att, err.Error())
				pconnObj.Close()
				return
			}
		}
	}(connObj)

	if err := nh.WebsocketReceiveNotification(connObj); err != nil {
		glog.Warningf("in ConnectToMaster. attributes: %s, reason: %s", cautils.ObjectToString(att), err.Error())
	}

	nh.wa.Close(connObj)
	if retry < 2 {
		glog.Warningf("disconnected from master with connection attributes: '%s', retrying: %d", cautils.ObjectToString(att), retry+1)
		nh.outgoingConnectionsMutex.Lock()
		nh.outgoingConnections.Remove(notificationAtt)
		nh.outgoingConnectionsMutex.Unlock()
		nh.ConnectToMaster(notificationAtt, retry+1)
	} else {
		glog.Warningf("disconnected from master with connection attributes: '%s', removing connection from list", cautils.ObjectToString(att))
		nh.outgoingConnectionsMutex.Lock()
		defer nh.outgoingConnectionsMutex.Unlock()

		nh.CleanupOutgoingConnection(att)
		if nh.outgoingConnections.Len() == 0 && nh.incomingConnections.Len() > 0 {
			panic(fmt.Sprintf("failed to connect to master: '%s'", cautils.ObjectToString(att)))
		}
	}
}

// RestAPINotificationHandler -
func (nh *NotificationServer) RestAPINotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		glog.Errorf("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", 405)
		return
	}

	readBuffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("In RestAPINotificationHandler ReadAll %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// get notificationID from message
	notificationAtt, err := nh.UnmarshalMessage(readBuffer)
	if err != nil {
		glog.Errorf("In RestAPINotificationHandler UnmarshalMessage %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	glog.Infof("REST API received message, attributes: %s", cautils.ObjectToString(notificationAtt.Target))
	if notificationAtt.Target == nil || len(notificationAtt.Target) == 0 {
		glog.Errorf("In RestAPINotificationHandler received empty notificationAtt.Target")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ids, err := nh.SendNotification(notificationAtt.Target, readBuffer, notificationAtt.SendSynchronicity)
	if err != nil {
		glog.Errorf("In RestAPINotificationHandler SendNotification %v, target: %v", err, notificationAtt.Target)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	byteIDs, _ := json.Marshal(ids)
	w.Write(byteIDs)

}

// SendNotification -
func (nh *NotificationServer) SendNotification(route map[string]string, notification []byte, sendSynchronicity bool) ([]int, error) {

	ids := []int{}
	errMsgs := []string{}
	connections := nh.incomingConnections.Get(route)
	glog.Infof("sending notification to: %v, number of connections: %d", cautils.ObjectToString(route), len(connections))
	if len(connections) == 0 {
		return ids, nil
	}
	preparedMessage, err := websocket.NewPreparedMessage(websocket.BinaryMessage, notification)
	if err != nil {
		return ids, fmt.Errorf("failed to prepare message, reason: %s", err.Error())
	}
	for _, conn := range connections {
		if sendSynchronicity {
			if err := nh.sendSingleNotification(conn, preparedMessage, 0); err != nil {
				errMsgs = append(errMsgs, err.Error())
			}
		} else {
			go nh.sendSingleNotification(conn, preparedMessage, 0)
		}
	}

	if len(errMsgs) > 0 {
		return ids, fmt.Errorf("%s", strings.Join(errMsgs, ";\n"))
	}
	return ids, nil
}
func (nh *NotificationServer) sendSingleNotification(conn *websocketactions.Connection, preparedMessage *websocket.PreparedMessage, retry int) error {
	defer func() {
		if err := recover(); err != nil {
			if retry < 2 && strings.Contains(fmt.Sprintf("%v", err), "concurrent write to websocket connection") {
				timeWait := time.Duration(rand.Intn(120)) * time.Millisecond
				glog.Errorf("recover sendSingleNotification, connection %d is not alive, error: %v, retry: %d, retrying in: %s", conn.ID, err, retry+1, timeWait.String())
				time.Sleep(timeWait)
				nh.sendSingleNotification(conn, preparedMessage, retry+1)
			} else {
				glog.Errorf("recover sendSingleNotification, connection %d is not alive, error: %v, closing connection", conn.ID, err)
				nh.wa.Close(conn)
			}
		}
	}()
	glog.Infof("sending notification, attributes: %s, id: %d", cautils.ObjectToString(conn.GetAttributes()), conn.ID)
	err := nh.wa.WritePreparedMessage(conn, preparedMessage)
	if err != nil {
		nh.CleanupIncomeConnection(conn.ID)
		e := fmt.Errorf("In sendSingleNotification %s, connection %d is not alive, error: %v", cautils.ObjectToString(conn.GetAttributes()), conn.ID, err)
		glog.Errorf(e.Error())
		return e
	}
	// glog.Infof("notification sent successfully, attributes: %v, id: %d", conn.GetAttributes(), conn.ID)
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

// CleanupIncomeConnection - cleanup one connection with matching id
func (nh *NotificationServer) CleanupIncomeConnection(id int) {
	// remove connection from list
	nh.incomingConnections.RemoveID(id)
}

// CleanupOutgoingConnection -
func (nh *NotificationServer) CleanupOutgoingConnection(notificationAtt map[string]string) {
	// remove outgoing connection from list
	nh.outgoingConnections.Remove(notificationAtt)

	// close all incoming connections related to this attributes
	glog.Infof("Removing master connection. Removing all incoming connections with attributes: '%s'", cautils.ObjectToString(notificationAtt))
	nh.incomingConnections.CloseConnections(nh.wa, notificationAtt)
}

// WebsocketReceiveNotification maintain websocket connection // RestAPIReceiveNotification -
func (nh *NotificationServer) WebsocketReceiveNotification(connObj *websocketactions.Connection) error {
	// Websocket ping pong
	for {
		msgType, message, err := nh.wa.ReadMessage(connObj)
		if err != nil {
			return err
		}
		// glog.Infof("In WebsocketReceiveNotification received msgType: %d. (text=1, close=8, ping=9)", msgType)
		switch msgType {
		case websocket.CloseMessage:
			return fmt.Errorf("In WebsocketReceiveNotification websocket received CloseMessage")
		case websocket.PingMessage:
			err := nh.wa.WritePongMessage(connObj)
			if err != nil {
				return fmt.Errorf("In WebsocketReceiveNotification WritePongMessage error: %v", err)
			}
		case websocket.TextMessage:

		case websocket.BinaryMessage:

		default:
			glog.Warningf("unknown message type")
			return nil
		}
		// get notificationID from message
		n, err := nh.UnmarshalMessage(message)
		if err != nil {
			glog.Errorf("In WebsocketReceiveNotification UnmarshalMessage error: %v", err)
			return fmt.Errorf("In WebsocketReceiveNotification UnmarshalMessage error: %v", err)
		}
		if n.Target == nil || len(n.Target) == 0 {
			glog.Errorf("In WebsocketReceiveNotification received empty notification.Target")
			return fmt.Errorf("In WebsocketReceiveNotification received empty notification.Target")
		}
		// send message
		if _, err := nh.SendNotification(n.Target, message, n.SendSynchronicity); err != nil {
			glog.Errorf("In WebsocketReceiveNotification SendNotification error: %v", err)
			return fmt.Errorf("In WebsocketReceiveNotification SendNotification error: %v", err)
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
