package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

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
			fmt.Printf("%s, Removing notification", notificationID)
			notificationMap[notificationID] = remove(notificationMap[notificationID], index)
			return
		}
	}
}

func waitForNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fmt.Printf("Method not allowed")
		http.Error(w, "Method not allowed", 405)
		return
	}

	notificationID := strings.Split(r.URL.Path, "/")[2]

	log.Printf("%s, Requesting notification", notificationID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	defer conn.Close()
	defer r.Body.Close()

	log.Printf("%s, connected successfully", notificationID)
	notificationMap[notificationID] = append(notificationMap[notificationID], conn)

	// Websocket ping pong
	for {
		msgType, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s, read Error: %v", notificationID, err)
			defer cleanupConnection(notificationID, conn)
			break
		}

		switch msgType {
		case websocket.CloseMessage:
			log.Printf("%s, Connection closed", notificationID)
			defer cleanupConnection(notificationID, conn)
			break
		case websocket.PingMessage:
			log.Printf("%s, Ping", notificationID)
			err = conn.WriteMessage(websocket.PongMessage, []byte("pong"))
			if err != nil {
				log.Printf("%s, Write Error: %v", notificationID, err)
				defer cleanupConnection(notificationID, conn)

			}
		}
	}
}

func sendNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Method not allowed. returning 405")
		http.Error(w, "Method not allowed", 405)
		return
	}

	notificationID := strings.Split(r.URL.Path, "/")[2]
	defer r.Body.Close()
	readBuffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	if err := sendNotification(notificationID, readBuffer); err != nil {
		log.Print(err)
		http.NotFound(w, r)
	}

}
func sendNotification(notificationID string, notification []byte) error {
	if _, ok := notificationMap[notificationID]; !ok {
		return fmt.Errorf("%s, notificationID not found", notificationID)
	}
	log.Printf("%s, Posting notification", notificationID)

	for _, connection := range notificationMap[notificationID] {
		s := string(notification)
		log.Printf("%s:\n%v", notificationID, s)
		err := connection.WriteMessage(websocket.TextMessage, notification)
		if err != nil {
			// Remove connection
			log.Printf("%s, connection %p is not alive", notificationID, connection)
			defer cleanupConnection(notificationID, connection)
		}
	}

	return nil

}

// func setNotificationHandle() error {

// 	// load configuration file
// 	// configURL := ""
// 	scheme := ""
// 	host := ""
// 	path := ""
// 	// notificationIDs := ""
// 	u := url.URL{Scheme: scheme, Host: host, Path: path, ForceQuery: false}

// 	// Websocket ping pong
// 	for {
// 		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 		// notificationMap[notificationID] = append(notificationMap[notificationID], conn)
// 		if err != nil {
// 			log.Printf("Error connecting to postman. url: %s\nMessage %#v", u.String(), err)
// 		}
// 		defer conn.Close()

// 		// Websocket receive message
// 		msgType, notification, err := conn.ReadMessage()
// 		if err != nil {
// 			// log.Printf("%s, read Error: %v", notificationID, err)
// 			// defer cleanupConnection(notificationID, conn)
// 			continue
// 		}

// 		switch msgType {
// 		case websocket.CloseMessage:
// 			// log.Printf("%s, Connection closed", notificationID)
// 			// defer cleanupConnection(notificationID, conn)
// 		case websocket.PingMessage:
// 			log.Printf("%s, Ping", host)
// 			err = conn.WriteMessage(websocket.PongMessage, []byte("pong"))
// 			if err != nil {
// 				log.Printf("%s, Write Error: %v", host, err)
// 				// defer cleanupConnection(notificationID, conn)
// 			}
// 		case websocket.TextMessage:
// 			// get notificationID from message
// 			notificationID := ""

// 			// send message
// 			sendNotification(notificationID, notification)

// 		case websocket.BinaryMessage:
// 			break
// 		}
// 	}

// }
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

	// go func() {
	// 	log.Fatal(setNotificationHandle())
	// }()

	<-finish
}
