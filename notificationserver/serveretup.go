package notificationserver

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	logger "github.com/dwertent/go-logger"
	"github.com/dwertent/go-logger/helpers"
)

var (
	PortRestAPI   = "8002"
	PortWebsocket = "8001"
)

// SetupNotificationServer set up listening http servers
func (ns *NotificationServer) SetupNotificationServer() {
	if port, ok := os.LookupEnv("CA_NOTIFICATION_SERVER_WS_PORT"); ok {
		PortWebsocket = port
	}
	if port, ok := os.LookupEnv("CA_NOTIFICATION_SERVER_PORT"); ok {
		PortRestAPI = port
	}
	finish := make(chan bool)

	server8002 := http.NewServeMux()
	var h8002 = new(RegexpHandler)
	r8002, _ := regexp.Compile(fmt.Sprintf("%s.*", notifier.PathRESTV1))
	h8002.HandleFunc(r8002, ns.RestAPINotificationHandler)
	server8002.Handle("/", h8002)
	go func() {
		logger.L().Fatal("", helpers.Error(http.ListenAndServe(fmt.Sprintf(":%s", PortRestAPI), server8002)))
	}()

	server8001 := http.NewServeMux()
	var h8001 = new(RegexpHandler)
	r8001, _ := regexp.Compile(fmt.Sprintf("%s.*", notifier.PathWebsocketV1))
	h8001.HandleFunc(r8001, ns.WebsocketNotificationHandler)
	server8001.Handle("/", h8001)
	go func() {
		logger.L().Fatal("", helpers.Error(http.ListenAndServe(fmt.Sprintf(":%s", PortWebsocket), server8001)))
	}()

	<-finish
}

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
