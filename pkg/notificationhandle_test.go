package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	"github.com/kubescape/gateway/pkg/websocketactions"
	"github.com/stretchr/testify/assert"
)

// NewNotificationServerMasterMock -
func NewNotificationServerMasterMock() *Gateway {
	return &Gateway{
		wa:                  &websocketactions.WebsocketActionsMock{},
		outgoingConnections: *NewConnectionsObj(),
		incomingConnections: *NewConnectionsObj(),
	}
}

// NewNotificationServerEdgeMock -
func NewNotificationServerEdgeMock() *Gateway {
	return &Gateway{
		wa:                  &websocketactions.WebsocketActionsMock{},
		outgoingConnections: *NewConnectionsObj(),
		incomingConnections: *NewConnectionsObj(),
	}
}

func HTTPRequestMock() *http.Request {
	r := &http.Request{}
	r.Method = http.MethodGet
	r.URL = &url.URL{Scheme: "https", Host: "localhost", Path: "v1/somepath", RawQuery: "customer=test&cluster=kube"}
	r.Body = &io.PipeReader{}
	return r
}

func TestParseURLQuery(t *testing.T) {
	ns := NewNotificationServerEdgeMock()
	att, err := ns.parseURLPath(&url.URL{RawQuery: fmt.Sprintf("%s=test&cluster=kube", notifier.TargetCustomer)})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(att))

	assert.Equal(t, att[notifier.TargetCustomer], "test")
	assert.Equal(t, att["cluster"], "kube")
}
