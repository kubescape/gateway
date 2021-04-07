package notificationserver

import (
	"notification-server/notificationserver/websocketactions"
	"sync"
	"testing"
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
	if len(rtv1) != 1 {
		t.Errorf("%v", rtv1)
	}

	att2 := map[string]string{"customer": "test"}
	rtv2 := cs.Get(att2)
	if len(rtv2) != 1 {
		t.Errorf("%v", rtv2)
	}

	att3 := map[string]string{"customer": "test", "cluster": "bla"}
	rtv3 := cs.Get(att3)
	if len(rtv3) != 0 {
		t.Errorf("%v", rtv3)
	}

	att4 := map[string]string{"cluster": "yay"}
	rtv4 := cs.Get(att4)
	if len(rtv4) != 1 {
		t.Errorf("%v", rtv4)
	}

	att5 := map[string]string{"customerGUID": "test"}
	rtv5 := cs.Get(att5)
	if len(rtv5) != 0 {
		t.Errorf("%v", rtv5)
	}
}
