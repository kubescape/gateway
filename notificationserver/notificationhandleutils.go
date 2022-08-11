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
	// RootHost -
	RootHost string
)

func SetupMasterInfo() {
	host, k0 := os.LookupEnv(MasterGatewayHostEnvironmentVariable)
	if !k0 {
		logger.L().Info("Running notification server as master")
		return
	}
	RootHost = host
	RootAttributes = []string{notifier.TargetCustomer, "customer"} // agent uses customer

	logger.L().Info("master info", helpers.String("host", RootHost), helpers.String("attributes", strutils.ObjectToString(RootAttributes)))
}

// IsMaster is server master or edge
func IsMaster() bool {
	return RootHost == ""
}
