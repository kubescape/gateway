package notificationserver

import (
	strutils "github.com/armosec/utils-go/str"
	"os"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
	"github.com/golang/glog"
)

var (
	// RootAttributes attributes the root component is expecting
	RootAttributes []string
	// MASTER_HOST -
	MASTER_HOST string
)

func SetupMasterInfo() {
	// att, k1 := os.LookupEnv("MASTER_NOTIFICATION_SERVER_ATTRIBUTES")
	host, k0 := os.LookupEnv("MASTER_NOTIFICATION_SERVER_HOST")
	if !k0 {
		glog.Infof("Running notification server as master")
		return
	}
	MASTER_HOST = host
	RootAttributes = []string{notifier.TargetCustomer, "customer"} // agent uses customer

	glog.Infof("master host: %s, master attributes: %v", MASTER_HOST, strutils.ObjectToString(RootAttributes))
}

// IsMaster is server master or edge
func IsMaster() bool {
	return MASTER_HOST == ""
}
