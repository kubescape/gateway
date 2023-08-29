package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	strutils "github.com/armosec/utils-go/str"
	logger "github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	"github.com/gorilla/websocket"
	beClientV1 "github.com/kubescape/backend/pkg/client/v1"
	"github.com/kubescape/backend/pkg/servicediscovery"
	v1 "github.com/kubescape/backend/pkg/servicediscovery/v1"
	"github.com/kubescape/gateway/pkg/websocketactions"
	"gopkg.in/mgo.v2/bson"
)

const serviceDiscoveryConfigPath = "/etc/config/services.json"

// Gateway is the main Gateway service object.
// It acts as a facade that manages incoming and outgoing connections, routes
// messages to recipients etc.
type Gateway struct {
	wa                       websocketactions.IWebsocketActions
	outgoingConnections      Connections
	incomingConnections      Connections
	outgoingConnectionsMutex *sync.Mutex
	rootGatewayURL           string
}

// NewGateway creates a new Gateway
func NewGateway() *Gateway {

	rootGatewayUrl := getRootGwUrl()

	return &Gateway{
		wa:                       websocketactions.NewWebsocketActions(),
		outgoingConnections:      *NewConnectionsObj(),
		incomingConnections:      *NewConnectionsObj(),
		outgoingConnectionsMutex: &sync.Mutex{},
		rootGatewayURL:           rootGatewayUrl,
	}
}

type Notification notifier.Notification

// WebsocketNotificationHandler establishes a websocket connection and handles incoming notifications
func (nh *Gateway) WebsocketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	// ----------------------------------------------------- 1
	// receive websocket connection from client
	if r.Method != http.MethodGet {
		logger.L().Error("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	go nh.connectToMaster(notificationAtt, 0)

	// ----------------------------------------------------- 4
	// Websocket read messages
	nh.WebsocketReceiveNotification(newConn)
	nh.CleanupIncomingConnection(id)
	nh.wa.Close(newConn)
}

// ConnectToMaster registers an incoming connection with given attributes with the Master Gateway
func (nh *Gateway) connectToMaster(notificationAtt map[string]string, retry int) {
	if nh.hasParent() { // only edge connects to master
		return
	}

	att := strutils.MergeSliceAndMap([]string{notifier.TargetCustomer}, notificationAtt)
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
	parentURL, err := beClientV1.GetRootGatewayUrl(nh.rootGatewayURL)
	if err != nil {
		logger.L().Error(err.Error())
		return
	}

	q := parentURL.Query()
	for i, j := range att {
		q.Add(i, j)
	}
	parentURL.RawQuery = q.Encode()
	logger.L().Info("connecting to master", helpers.String("url", parentURL.String()))

	// connect to master
	conn, _, err := nh.wa.DefaultDialer(parentURL.String())
	if err != nil {
		logger.L().Fatal("failed to connect to master", helpers.String("url", parentURL.String()), helpers.Error(err))
	}
	connObj, _ := nh.outgoingConnections.Append(att, conn)
	nh.outgoingConnectionsMutex.Unlock()

	logger.L().Info("successfully contented to master", helpers.Int("number of outgoing websockets", nh.outgoingConnections.Len()))

	// read/write for keeping websocket connection alive
	go func(pconnObj *websocketactions.Connection) {
		for {
			time.Sleep(10 * time.Second)
			if err := nh.wa.WritePingMessage(pconnObj); err != nil {
				logger.L().Warning("in WritePingMessage", helpers.Interface("attributes", att), helpers.Error(err))
				pconnObj.Close()
				return
			}
		}
	}(connObj)

	if err := nh.WebsocketReceiveNotification(connObj); err != nil {
		logger.L().Warning("in connectToMaster", helpers.String("attributes", strutils.ObjectToString(att)), helpers.Error(err))
	}

	nh.wa.Close(connObj)
	if retry < 2 {
		logger.L().Warning("disconnected from master with connection", helpers.String("attributes", strutils.ObjectToString(att)), helpers.Int("retrying", retry+1))
		nh.outgoingConnectionsMutex.Lock()
		nh.outgoingConnections.Remove(notificationAtt)
		nh.outgoingConnectionsMutex.Unlock()
		nh.connectToMaster(notificationAtt, retry+1)
	} else {
		logger.L().Warning("disconnected from master with connection, removing connection from list", helpers.String("attributes", strutils.ObjectToString(att)))
		nh.outgoingConnectionsMutex.Lock()
		defer nh.outgoingConnectionsMutex.Unlock()

		nh.CleanupOutgoingConnection(att)
		if nh.outgoingConnections.Len() == 0 && nh.incomingConnections.Len() > 0 {
			logger.L().Fatal(fmt.Sprintf("failed to connect to parent: '%s'", strutils.ObjectToString(att)))
		}
	}
}

// RestAPINotificationHandler handles the notifications received over the REST API
func (nh *Gateway) RestAPINotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.L().Error("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	readBuffer, err := io.ReadAll(r.Body)
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

// SendNotification sends a notification to its intended recipient
func (nh *Gateway) SendNotification(route map[string]string, notification []byte, sendSynchronicity bool) ([]int, error) {

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

func (nh *Gateway) sendSingleNotification(conn *websocketactions.Connection, preparedMessage *websocket.PreparedMessage, retry int) error {
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
		nh.CleanupIncomingConnection(conn.ID)
		e := fmt.Errorf("in sendSingleNotification %s, connection %d is not alive, error: %v", strutils.ObjectToString(conn.GetAttributes()), conn.ID, err)
		logger.L().Error(e.Error())
		return e
	}
	return nil
}

// AcceptWebsocketConnection accepts an incoming websocket connection
func (nh *Gateway) AcceptWebsocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, map[string]string, error) {

	notificationAtt, err := nh.parseURLPath(r.URL)
	if err != nil {
		return nil, notificationAtt, err
	}

	conn, err := nh.wa.ConnectWebsocket(w, r)
	if err != nil {
		return conn, notificationAtt, err
	}

	return conn, notificationAtt, nil

}

// CleanupIncomingConnection cleans up an incoming connection with a given ID
func (nh *Gateway) CleanupIncomingConnection(id int) {
	// remove connection from list
	nh.incomingConnections.RemoveID(id)
}

// CleanupOutgoingConnection cleans up an incoming connection with given notification attributes
func (nh *Gateway) CleanupOutgoingConnection(notificationAtt map[string]string) {
	// remove outgoing connection from list
	nh.outgoingConnections.Remove(notificationAtt)

	// close all incoming connections related to this attributes
	logger.L().Info("Removing master connection. Removing all incoming connections", helpers.String("attributes", strutils.ObjectToString(notificationAtt)))
	nh.incomingConnections.CloseConnections(nh.wa, notificationAtt)
}

// WebsocketReceiveNotification maintains the websocket connection and receives notifications sent over it
func (nh *Gateway) WebsocketReceiveNotification(connObj *websocketactions.Connection) error {
	// Websocket ping pong
	for {
		msgType, message, err := nh.wa.ReadMessage(connObj)
		if err != nil {
			return err
		}
		switch msgType {
		case websocket.CloseMessage:
			return fmt.Errorf("in WebsocketReceiveNotification websocket received CloseMessage")
		case websocket.PingMessage:
			err := nh.wa.WritePongMessage(connObj)
			if err != nil {
				return fmt.Errorf("in WebsocketReceiveNotification WritePongMessage error: %v", err)
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
			logger.L().Error("in WebsocketReceiveNotification UnmarshalMessage", helpers.Error(err))
			return fmt.Errorf("in WebsocketReceiveNotification UnmarshalMessage error: %v", err)
		}
		if n.Target == nil || len(n.Target) == 0 {
			logger.L().Error("In WebsocketReceiveNotification received empty notification.Target")
			return fmt.Errorf("in WebsocketReceiveNotification received empty notification.Target")
		}
		// send message
		if _, err := nh.SendNotification(n.Target, message, n.SendSynchronicity); err != nil {
			logger.L().Error("In WebsocketReceiveNotification SendNotification", helpers.Error(err))
			return fmt.Errorf("in WebsocketReceiveNotification SendNotification error: %v", err)
		}
	}
}

// parseURLPath transforms a given URL path parameters to notification attributes
func (nh *Gateway) parseURLPath(u *url.URL) (map[string]string, error) {

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

// UnmarshalMessage attempts to unmarshal a given message into either a JSON or BSON format
func (nh *Gateway) UnmarshalMessage(message []byte) (*Notification, error) {
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

// hasParent does the parent host is set
func (nh *Gateway) hasParent() bool {
	return nh.rootGatewayURL == ""
}

// getRootGwUrl return the parent host URL. According to the following order: env var, service discovery, config file
func getRootGwUrl() string {
	if envVarValue := os.Getenv(ParentGatewayHostEnvironmentVariable); envVarValue != "" {
		logger.L().Info("loaded gw url from env var", helpers.String("url", envVarValue))
		return envVarValue
	}

	services, err := servicediscovery.GetServices(v1.NewServiceDiscoveryFileV1(serviceDiscoveryConfigPath))
	if err != nil {
		logger.L().Warning(err.Error())
	}
	url := services.GetGatewayUrl()
	logger.L().Info("loaded gw url (service discovery)", helpers.String("url", url))

	return url
}
