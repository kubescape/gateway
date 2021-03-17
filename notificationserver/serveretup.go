package notificationserver

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"asterix.cyberarmor.io/cyberarmor/capacketsgo/notificationserver"
	"github.com/golang/glog"
)

var (
	MASTER_REST_PORT = "8002"
	WEBSOCKET_PORT   = "8001"
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
	if port, ok := os.LookupEnv("CA_WEBSOCKET_PORT"); ok {
		WEBSOCKET_PORT = port
	}
	if port, ok := os.LookupEnv("CA_REST_API_PORT"); ok {
		MASTER_REST_PORT = port
	}
	finish := make(chan bool)

	server8002 := http.NewServeMux()
	var h8002 = new(RegexpHandler)
	r8002, _ := regexp.Compile(fmt.Sprintf("%s.*", notificationserver.PathRESTV1))
	h8002.HandleFunc(r8002, ns.RestAPINotificationHandler)
	server8002.Handle("/", h8002)
	go func() {
		glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", MASTER_REST_PORT), server8002))
	}()

	server8001 := http.NewServeMux()
	var h8001 = new(RegexpHandler)
	r8001, _ := regexp.Compile(fmt.Sprintf("%s.*", notificationserver.PathWebsocketV1))
	h8001.HandleFunc(r8001, ns.WebsocketNotificationHandler)
	server8001.Handle("/", h8001)
	go func() {
		glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", WEBSOCKET_PORT), server8001))
	}()

	<-finish
}
