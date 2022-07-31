package notificationserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	strutils "github.com/armosec/utils-go/str"
	logger "github.com/dwertent/go-logger"
	"github.com/dwertent/go-logger/helpers"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	"github.com/armosec/utils-k8s-go/armometadata"
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
		logger.L().Warning(err.Error())
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
		logger.L().Error("Method not allowed")
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, notificationAtt, err := nh.AcceptWebsocketConnection(w, r)
	if err != nil {
		logger.L().Error(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	// ----------------------------------------------------- 2
	// append new route
	newConn, id := nh.incomingConnections.Append(notificationAtt, conn)
	logger.L().Info("accepting websocket connection", helpers.String("url query", r.URL.RawQuery), helpers.Int("id", id), helpers.Int("number of incoming websockets", nh.incomingConnections.Len()))

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

	att := strutils.MergeSliceAndMap(RootAttributes, notificationAtt)
	if len(att) == 0 {
		att = notificationAtt
	}
	nh.outgoingConnectionsMutex.Lock() // lock connecting to master to prevent many connections

	// if connected
	if cons := nh.outgoingConnections.Get(att); len(cons) > 0 {
		nh.outgoingConnectionsMutex.Unlock()
		logger.L().Info("edge already connected to master, not creating new connection")
		return
	}

	masterURL := fmt.Sprintf("%s?", MASTER_HOST)
	amp := ""
	for i, j := range att {
		masterURL += amp
		masterURL += fmt.Sprintf("%s=%s", i, j)
		amp = "&"
	}
	logger.L().Info("connecting to master", helpers.String("url", masterURL))

	// connect to master
	conn, _, err := nh.wa.DefaultDialer(masterURL, nil)
	if err != nil {
		nh.outgoingConnectionsMutex.Unlock()
		logger.L().Error("in ConnectToMaster", helpers.Error(err))
		return
	}
	connObj, _ := nh.outgoingConnections.Append(att, conn)
	nh.outgoingConnectionsMutex.Unlock()

	logger.L().Info("successfully contented to master", helpers.Int("number of outgoing websockets", nh.outgoingConnections.Len()))

	// read/write for keeping websocket connection alive
	go func(pconnObj *websocketactions.Connection) {
		for {
			time.Sleep(10 * time.Second)
			if err := nh.wa.WritePingMessage(pconnObj); err != nil {
				logger.L().Warning("in WritePingMessage", helpers.Interface(" attributes", att), helpers.Error(err))
				pconnObj.Close()
				return
			}
		}
	}(connObj)

	if err := nh.WebsocketReceiveNotification(connObj); err != nil {
		logger.L().Warning("in ConnectToMaster", helpers.String("attributes", strutils.ObjectToString(att)), helpers.Error(err))
	}

	nh.wa.Close(connObj)
	if retry < 2 {
		logger.L().Warning("disconnected from master with connection", helpers.String("attributes", strutils.ObjectToString(att)), helpers.Int("retrying", retry+1))
		nh.outgoingConnectionsMutex.Lock()
		nh.outgoingConnections.Remove(notificationAtt)
		nh.outgoingConnectionsMutex.Unlock()
		nh.ConnectToMaster(notificationAtt, retry+1)
	} else {
		logger.L().Warning("disconnected from master with connection, removing connection from list", helpers.String("attributes", strutils.ObjectToString(att)))
		nh.outgoingConnectionsMutex.Lock()
		defer nh.outgoingConnectionsMutex.Unlock()

		nh.CleanupOutgoingConnection(att)
		if nh.outgoingConnections.Len() == 0 && nh.incomingConnections.Len() > 0 {
			panic(fmt.Sprintf("failed to connect to master: '%s'", strutils.ObjectToString(att)))
		}
	}
}

// RestAPINotificationHandler -
func (nh *NotificationServer) RestAPINotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.L().Error("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", 405)
		return
	}

	readBuffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.L().Error("In RestAPINotificationHandler ReadAll", helpers.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// get notificationID from message
	notificationAtt, err := nh.UnmarshalMessage(readBuffer)
	if err != nil {
		logger.L().Error("in RestAPINotificationHandler UnmarshalMessage", helpers.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.L().Info("in RestAPINotificationHandler", helpers.String("attributes", strutils.ObjectToString(notificationAtt.Target)))
	if notificationAtt.Target == nil || len(notificationAtt.Target) == 0 {
		logger.L().Error("in RestAPINotificationHandler received empty notificationAtt.Target")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ids, err := nh.SendNotification(notificationAtt.Target, readBuffer, notificationAtt.SendSynchronicity)
	if err != nil {
		logger.L().Error("in RestAPINotificationHandler SendNotification", helpers.String("target", strutils.ObjectToString(notificationAtt.Target)), helpers.Error(err))
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
	logger.L().Info("sending notification", helpers.Interface("target", strutils.ObjectToString(route)), helpers.Int("number of connections", len(connections)))
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

				logger.L().Error("recover sendSingleNotification, connection is not alive", helpers.Int("id", conn.ID), helpers.Interface("reason", err), helpers.Int("retry", retry+1), helpers.String("retrying in", timeWait.String()))
				time.Sleep(timeWait)
				nh.sendSingleNotification(conn, preparedMessage, retry+1)
			} else {
				logger.L().Error("recover sendSingleNotification, connection is not alive", helpers.Int("id", conn.ID), helpers.Interface("reason", err))
				nh.wa.Close(conn)
			}
		}
	}()
	logger.L().Info("sending notification", helpers.String("attributes", strutils.ObjectToString(conn.GetAttributes())), helpers.Int("id", conn.ID))
	err := nh.wa.WritePreparedMessage(conn, preparedMessage)
	if err != nil {
		nh.CleanupIncomeConnection(conn.ID)
		e := fmt.Errorf("in sendSingleNotification %s, connection %d is not alive, error: %v", strutils.ObjectToString(conn.GetAttributes()), conn.ID, err)
		logger.L().Error(e.Error())
		return e
	}
	// logger.L().Info("notification sent successfully, attributes: %v, id: %d", conn.GetAttributes(), conn.ID)
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
	logger.L().Info("Removing master connection. Removing all incoming connections", helpers.String("attributes", strutils.ObjectToString(notificationAtt)))
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
		// logger.L().Info("In WebsocketReceiveNotification received msgType: %d. (text=1, close=8, ping=9)", msgType)
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
			logger.L().Warning("unknown message type")
			return nil
		}
		// get notificationID from message
		n, err := nh.UnmarshalMessage(message)
		if err != nil {
			logger.L().Error("In WebsocketReceiveNotification UnmarshalMessage", helpers.Error(err))
			return fmt.Errorf("In WebsocketReceiveNotification UnmarshalMessage error: %v", err)
		}
		if n.Target == nil || len(n.Target) == 0 {
			logger.L().Error("In WebsocketReceiveNotification received empty notification.Target")
			return fmt.Errorf("In WebsocketReceiveNotification received empty notification.Target")
		}
		// send message
		if _, err := nh.SendNotification(n.Target, message, n.SendSynchronicity); err != nil {
			logger.L().Error("In WebsocketReceiveNotification SendNotification", helpers.Error(err))
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
