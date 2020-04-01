package notificationserver

import (
	"sync"
	"testing"
)

var ATTRIBUTES_MOCK = map[string]string{"customer": "test", "cluster": "yay"}

func ConnectionMock() *Connection {
	return &Connection{
		attributes: ATTRIBUTES_MOCK,
	}

}
func ConnectionsMock() *Connections {
	return &Connections{
		connections: []*Connection{
			ConnectionMock(),
		},
		mutex: &sync.Mutex{},
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
}
