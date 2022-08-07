package notificationserver

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupMasterInfo(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{
			name: "localhost",
			host: "localhost",
		},
		{
			name: "other-host",
			host: "other-host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(MasterNotificationServerHostEnvironmentVariable, tt.host)
			SetupMasterInfo()
			assert.Equal(t, tt.host, RootHost)
		})
	}
}

func TestIsMaster(t *testing.T) {
	tests := []struct {
		name     string
		want     bool
		rootHost string
	}{
		{
			name:     "is master",
			want:     true,
			rootHost: "",
		},
		{
			name:     "not master",
			want:     false,
			rootHost: "localhost",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RootHost = tt.rootHost
			if got := IsMaster(); got != tt.want {
				t.Errorf("IsMaster() = %v, want %v", got, tt.want)
			}
		})
	}
}
