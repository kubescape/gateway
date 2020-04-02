package notificationserver

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
)

var (
	VERSION          = "v1"
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

	if IsMaster() {
		server8002 := http.NewServeMux()
		var h8002 = new(RegexpHandler)
		r8002, _ := regexp.Compile(fmt.Sprintf("/%s/%s.*", VERSION, MASTER_REST_API))
		h8002.HandleFunc(r8002, ns.RestAPINotificationHandler)
		server8002.Handle("/", h8002)
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", MASTER_REST_PORT), server8002))
		}()
	}

	server8001 := http.NewServeMux()
	var h8001 = new(RegexpHandler)
	r8001, _ := regexp.Compile(fmt.Sprintf("/%s/%s.*", VERSION, WEBSOCKET_API))
	h8001.HandleFunc(r8001, ns.WebsocketNotificationHandler)
	server8001.Handle("/", h8001)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", WEBSOCKET_PORT), server8001))
	}()

	<-finish
}
