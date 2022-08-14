package gateway

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/kubescape/gateway/docs"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	logger "github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
)

var (
	PortRestAPI   = "8002"
	PortWebsocket = "8001"
)

// SetupAndServe configures the HTTP servers and makes them serve incoming requests
func (ns *Gateway) SetupAndServe() {
	if port, ok := os.LookupEnv(GatewayWebsocketPortEnvironmentVariable); ok {
		PortWebsocket = port
	}
	if port, ok := os.LookupEnv(GatewayRestApiPortEnvironmentVariable); ok {
		PortRestAPI = port
	}
	finish := make(chan bool)

	restAPIServer := http.NewServeMux()
	var restAPIHandler = new(RegexpHandler)
	restAPIRoute, _ := regexp.Compile(fmt.Sprintf("%s.*", notifier.PathRESTV1))
	restAPIHandler.HandleFunc(restAPIRoute, ns.RestAPINotificationHandler)
	restAPIServer.Handle("/", restAPIHandler)

	openAPIHandler := docs.NewOpenAPIUIHandler()
	restAPIServer.Handle(docs.OpenAPIV2Prefix, openAPIHandler)

	go func() {
		logger.L().Fatal("", helpers.Error(http.ListenAndServe(fmt.Sprintf(":%s", PortRestAPI), restAPIServer)))
	}()

	websocketServer := http.NewServeMux()
	var websocketHandler = new(RegexpHandler)
	websocketRoute, _ := regexp.Compile(fmt.Sprintf("%s.*", notifier.PathWebsocketV1))
	websocketHandler.HandleFunc(websocketRoute, ns.WebsocketNotificationHandler)
	websocketServer.Handle("/", websocketHandler)
	go func() {
		logger.L().Fatal("", helpers.Error(http.ListenAndServe(fmt.Sprintf(":%s", PortWebsocket), websocketServer)))
	}()

	<-finish
}

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

// RegexpHandler describes a Handler that handles requests that have a URL path
// that matches an associated regular expression
type RegexpHandler struct {
	routes []*route
}

// Handler registers a given handler to the provided route pattern
func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

// HandleFunc registers a given handler function to the provided route pattern
func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

// ServeHTTP serves the handler
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
