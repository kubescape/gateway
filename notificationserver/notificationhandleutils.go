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

// SetupMasterInfo sets up the info about the Master Gateway
func SetupMasterInfo() {
	host, k0 := os.LookupEnv(MasterGatewayHostEnvironmentVariable)
	if !k0 {
		logger.L().Info("Running Gateway as master")
		return
	}
	RootHost = host
	RootAttributes = []string{notifier.TargetCustomer, "customer"} // agent uses customer

	logger.L().Info("master info", helpers.String("host", RootHost), helpers.String("attributes", strutils.ObjectToString(RootAttributes)))
}

// IsMaster checks if the current Gateway instance is a Master or an Edge
func IsMaster() bool {
	return RootHost == ""
}
