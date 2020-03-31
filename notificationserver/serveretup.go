package notificationserver

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
)

var (
	MASTER_REST_API  = "sendnotification"
	MASTER_REST_PORT = 8002
	WEBSOCKET_API    = "waitfornotification"
	WEBSOCKET_PORT   = 8001
)

type notificationHandlerFunc func(http.ResponseWriter, *http.Request)

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

// SetupNotificationServer set up listening http servers
func (ns *NotificationServer) SetupNotificationServer() {
	finish := make(chan bool)
	log.Printf("SetupNotificationServer")
	if IsMaster() {
		log.Printf("1 SetupNotificationHandler")
		SetupNotificationHandler(ns.RestAPINotificationHandler, MASTER_REST_API, MASTER_REST_PORT)
		log.Printf("2 SetupNotificationHandler")

	}
	log.Printf("3 SetupNotificationHandler")
	SetupNotificationHandler(ns.WebsocketNotificationHandler, WEBSOCKET_API, WEBSOCKET_PORT)
	log.Printf("4 SetupNotificationHandler")

	<-finish

}

// SetupNotificationHandler set up listening websocket/restAPI
func SetupNotificationHandler(handler notificationHandlerFunc, api string, port int) {
	log.Printf("listening %d:%s", port, api)

	rCompile, _ := regexp.Compile(fmt.Sprintf("/%s.+", api))

	var regexpHandler = new(RegexpHandler)
	regexpHandler.HandleFunc(rCompile, handler)

	serverHandler := http.NewServeMux()
	serverHandler.Handle("/", regexpHandler)

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), serverHandler))
	}()

}
