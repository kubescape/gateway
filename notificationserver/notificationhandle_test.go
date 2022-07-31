package notificationserver

import (
	"io"
	"net/http"
	"net/url"
	"notification-server/notificationserver/websocketactions"
	"testing"
)

// NewNotificationServerMasterMock -
func NewNotificationServerMasterMock() *NotificationServer {
	return &NotificationServer{
		wa:                  &websocketactions.WebsocketActionsMock{},
		outgoingConnections: *NewConnectionsObj(),
		incomingConnections: *NewConnectionsObj(),
	}
}

// NewNotificationServerEdgeMock -
func NewNotificationServerEdgeMock() *NotificationServer {
	MASTER_HOST = "https://blabla"
	RootAttributes = []string{"customer"}

	return &NotificationServer{
		wa:                  &websocketactions.WebsocketActionsMock{},
		outgoingConnections: *NewConnectionsObj(),
		incomingConnections: *NewConnectionsObj(),
	}
}

func HTTPRequestMock() *http.Request {
	r := &http.Request{}
	r.Method = http.MethodGet
	r.URL = &url.URL{Scheme: "https://", Host: "blabla", Path: "v1/somepath", RawQuery: "customer=test&cluster=kube"}
	r.Body = &io.PipeReader{}
	return r
}

// func TestWebsocketNotificationHandlerMaster(t *testing.T) {
// 	nsm := NewNotificationServerMasterMock()

// 	r := HTTPRequestMock()
// 	nsm.WebsocketNotificationHandler(nil, r)

// 	if len(nsm.outgoingConnections.connections) > 0 || len(nsm.incomingConnections.connections) > 0 {
// 		t.Error("connections where not removed")
// 	}
// }

func TestParseURLQuery(t *testing.T) {
	ns := NewNotificationServerEdgeMock()
	att, err := ns.ParseURLPath(&url.URL{RawQuery: "customer=test&cluster=kube"})
	if err != nil {
		t.Error(err)
	}
	if len(att) != 2 {
		t.Error("len(att) != 2")
	}
	if att["customer"] != "test" || att["cluster"] != "kube" {
		t.Error("wrong key value")
	}
}

func TestParseURLPath(t *testing.T) {
	ns := NewNotificationServerEdgeMock()
	att, err := ns.ParseURLPath(&url.URL{Path: "/v1/notify/my-id"})
	if err != nil {
		t.Error(err)
	}
	if len(att) != 1 {
		t.Error("len(att) != 1")
	}
	if att["my-id"] != "" {
		t.Error("wrong key value")
	}
}
