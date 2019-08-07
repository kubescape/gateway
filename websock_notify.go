package main

import (
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

func cleanupConnection(notificationId string, conn *websocket.Conn) {
	for index, element := range notificationMap[notificationId] {
		if element == conn {
			notificationMap[notificationId] = remove(notificationMap[notificationId], index)
			return
		}
	}
}

func waitForNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	notificationId := strings.Split(r.URL.Path, "/")[2]

	log.Printf("Requesting notification for %s", notificationId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	notificationMap[notificationId] = append(notificationMap[notificationId], conn)
}

func sendNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	notificationId := strings.Split(r.URL.Path, "/")[2]

	if _, ok := notificationMap[notificationId]; ok {
		log.Printf("Posting notification for %s", notificationId)
		defer r.Body.Close()
		readBuffer, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			return
		}
		for {
			for _, connection := range notificationMap[notificationId] {
				s := string(readBuffer)
				log.Print(s)
				err = connection.WriteMessage(websocket.TextMessage, readBuffer)
				if err != nil {
					// Remove connection
					log.Printf("connection %p is not alive", connection)
					defer cleanupConnection(notificationId, connection)
				}
			}
		}
	} else {
		http.NotFound(w, r)
	}

}

func main() {
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
