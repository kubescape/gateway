package notificationserver

import (
	"os"

	strutils "github.com/armosec/utils-go/str"
	logger "github.com/dwertent/go-logger"
	"github.com/dwertent/go-logger/helpers"

	notifier "github.com/armosec/cluster-notifier-api-go/notificationserver"
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
		logger.L().Info("Running notification server as master")
		return
	}
	MASTER_HOST = host
	RootAttributes = []string{notifier.TargetCustomer, "customer"} // agent uses customer

	logger.L().Info("master info", helpers.String("host", MASTER_HOST), helpers.String("attributes", strutils.ObjectToString(RootAttributes)))
}

// IsMaster is server master or edge
func IsMaster() bool {
	return MASTER_HOST == ""
}
