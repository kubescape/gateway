package websocketactions

import (
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestConnection_AttributesContained(t *testing.T) {
	type fields struct {
		mutex      *sync.Mutex
		ID         int
		conn       *websocket.Conn
		attributes map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		search map[string]string
		want   bool
	}{
		{
			fields: fields{
				attributes: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			search: map[string]string{
				"a": "b",
				"c": "d",
			},
			want: true,
			name: "contained all attributes",
		},
		{
			fields: fields{
				attributes: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			search: map[string]string{
				"c": "d",
			},
			want: true,
			name: "contained some attributes",
		},
		{
			fields: fields{
				attributes: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			search: map[string]string{
				"v": "d",
			},
			name: "does not contain attributes",
			want: false,
		},
		{
			fields: fields{
				attributes: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			search: map[string]string{},
			name:   "does not contain any attributes",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Connection{
				mutex:      tt.fields.mutex,
				ID:         tt.fields.ID,
				conn:       tt.fields.conn,
				attributes: tt.fields.attributes,
			}
			if got := c.AttributesContained(tt.search); got != tt.want {
				t.Errorf("Connection.AttributesContained() = %v, want %v", got, tt.want)
			}
		})
	}
}
