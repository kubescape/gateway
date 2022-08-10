package notificationserver

import (
	"sync"
	"testing"

	"github.com/kubescape/gateway/notificationserver/websocketactions"

	"github.com/stretchr/testify/assert"
)

var ATTRIBUTES_MOCK = map[string]string{"customer": "test", "cluster": "yay"}

func ConnectionMock() *websocketactions.Connection {
	return websocketactions.NewConnection(nil, 1234, ATTRIBUTES_MOCK)
}
func ConnectionsMock() *Connections {
	return &Connections{
		connections: []*websocketactions.Connection{
			ConnectionMock(),
		},
		mutex: &sync.RWMutex{},
	}
}
func TestGet(t *testing.T) {
	cs := ConnectionsMock()

	att1 := ATTRIBUTES_MOCK
	rtv1 := cs.Get(att1)
	assert.Equal(t, 1, len(rtv1))

	att2 := map[string]string{"customer": "test"}
	rtv2 := cs.Get(att2)
	assert.Equal(t, 1, len(rtv2))

	att3 := map[string]string{"customer": "test", "cluster": "bla"}
	rtv3 := cs.Get(att3)
	assert.Equal(t, 0, len(rtv3))

	att4 := map[string]string{"cluster": "yay"}
	rtv4 := cs.Get(att4)
	assert.Equal(t, 1, len(rtv4))

	att5 := map[string]string{"customerGUID": "test"}
	rtv5 := cs.Get(att5)
	assert.Equal(t, 0, len(rtv5))

}
