package notificationserver

import (
	"io"
	"net/http"
	"net/url"
	"testing"

	"canotificationserver/notificationserver/websocketactions"
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
	MASTER_ATTRIBUTES = []string{"customer"}

	return &NotificationServer{
		wa:                  &websocketactions.WebsocketActionsMock{},
		outgoingConnections: *NewConnectionsObj(),
		incomingConnections: *NewConnectionsObj(),
	}
}

func HTTPRequestMock() *http.Request {
	r := &http.Request{}
	r.Method = http.MethodGet
	r.URL = &url.URL{Scheme: "https://", Host: "blabla", Path: "somepath", RawQuery: "customer=test&cluster=kube"}
	r.Body = &io.PipeReader{}
	return r
}
func TestWebsocketNotificationHandlerMaster(t *testing.T) {
	nsm := NewNotificationServerMasterMock()

	r := HTTPRequestMock()
	nsm.WebsocketNotificationHandler(nil, r)

	if len(nsm.outgoingConnections.connections) > 0 || len(nsm.incomingConnections.connections) > 0 {
		t.Errorf("connections where not removed")
	}
}

func TestParseURLQuery(t *testing.T) {
	ns := NewNotificationServerEdgeMock()
	att, err := ns.ParseURLPath(&url.URL{RawQuery: "customer=test&cluster=kube"})
	if err != nil {
		t.Error(err)
	}
	if len(att) != 2 {
		t.Errorf("len(att) != 2")
	}
	if att["customer"] != "test" || att["cluster"] != "kube" {
		t.Errorf("worng key value")
	}
}

func TestParseURLPath(t *testing.T) {
	ns := NewNotificationServerEdgeMock()
	att, err := ns.ParseURLPath(&url.URL{Path: "/notify/my-id"})
	if err != nil {
		t.Error(err)
	}
	if len(att) != 1 {
		t.Errorf("len(att) != 1")
	}
	if att["my-id"] != "" {
		t.Errorf("worng key value")
	}
}
