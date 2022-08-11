package main

import (
	"flag"
	"os"

	"github.com/kubescape/gateway/notificationserver"

	"github.com/armosec/utils-k8s-go/probes"
	logger "github.com/dwertent/go-logger"
	"github.com/dwertent/go-logger/helpers"
)

//go:generate swagger generate spec -o ./docs/swagger.yaml
func main() {
	flag.Parse()

	displayBuildTag()

	isReadinessReady := false
	go probes.InitReadinessV1(&isReadinessReady)

	ns := notificationserver.NewGateway()
	isReadinessReady = true
	ns.SetupNotificationServer()
}

// DisplayBuildTag display on startup
func displayBuildTag() {
	logger.L().Info("Image version", helpers.String("release", os.Getenv(notificationserver.ReleaseBuildTagEnvironmentVariable)))
}
