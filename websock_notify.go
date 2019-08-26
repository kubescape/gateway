package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

var notificationMap map[string][]*websocket.Conn

func remove(s []*websocket.Conn, i int) []*websocket.Conn {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func cleanupConnection(notificationID string, conn *websocket.Conn) {
	for index, element := range notificationMap[notificationID] {
		if element == conn {
			glog.Infof("%s, Removing notification", notificationID)
			notificationMap[notificationID] = remove(notificationMap[notificationID], index)
			return
		}
	}
}

func waitForNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		glog.Errorf("Method not allowed")
		http.Error(w, "Method not allowed", 405)
		return
	}

	notificationID := strings.Split(r.URL.Path, "/")[2]

	glog.Infof("%s, Requesting notification", notificationID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("%v", err)
		return
	}
	defer conn.Close()
	defer r.Body.Close()

	glog.Infof("%s, connected successfully", notificationID)
	notificationMap[notificationID] = append(notificationMap[notificationID], conn)

	// Websocket ping pong
	for {
		msgType, _, err := conn.ReadMessage()
		if err != nil {
			glog.Errorf("%s, read Error: %v", notificationID, err)
			defer cleanupConnection(notificationID, conn)
			break
		}

		switch msgType {
		case websocket.CloseMessage:
			glog.Errorf("%s, Connection closed", notificationID)
			defer cleanupConnection(notificationID, conn)
			break
		case websocket.PingMessage:
			glog.Infof("%s, Ping", notificationID)
			err = conn.WriteMessage(websocket.PongMessage, []byte("pong"))
			if err != nil {
				glog.Errorf("%s, Write Error: %v", notificationID, err)
				defer cleanupConnection(notificationID, conn)

			}
		}
	}
}

func sendNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		glog.Errorf("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", 405)
		return
	}

	notificationID := strings.Split(r.URL.Path, "/")[2]

	if _, ok := notificationMap[notificationID]; ok {
		glog.Infof("%s, Posting notification", notificationID)
		defer r.Body.Close()
		readBuffer, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Errorf("%v", err)
			return
		}
		for _, connection := range notificationMap[notificationID] {
			s := string(readBuffer)
			glog.Infof("%s:\n%v", notificationID, s)
			err = connection.WriteMessage(websocket.TextMessage, readBuffer)
			if err != nil {
				// Remove connection
				glog.Errorf("%s, connection %p is not alive", notificationID, connection)
				defer cleanupConnection(notificationID, connection)
			}
		}
	} else {
		glog.Errorf("%s, notificationID not found", notificationID)
		http.NotFound(w, r)
	}

}

func main() {
	flag.Parse()

	notificationMap = make(map[string][]*websocket.Conn)

	finish := make(chan bool)

	server8001 := http.NewServeMux()
	server8002 := http.NewServeMux()

	var h8001 = new(RegexpHandler)
	var h8002 = new(RegexpHandler)

	r8001, _ := regexp.Compile("/waitfornotification/.+")
	r8002, _ := regexp.Compile("/sendnotification/.+")

	h8001.HandleFunc(r8001, waitForNotificationHandle)
	h8002.HandleFunc(r8002, sendNotificationHandle)

	server8001.Handle("/", h8001)
	server8002.Handle("/", h8002)

	go func() {
		log.Fatal(http.ListenAndServe(":8001", server8001))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8002", server8002))
	}()

	<-finish
}
