package main

import (
	"flag"
	"notification-server/notificationserver"
	"os"

	"github.com/armosec/utils-k8s-go/probes"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "1")

	displayBuildTag()

	isReadinessReady := false
	go probes.InitReadinessV1(&isReadinessReady)

	ns := notificationserver.NewNotificationServer()
	isReadinessReady = true
	ns.SetupNotificationServer()
}

// DisplayBuildTag display on startup
func displayBuildTag() {
	glog.Infof("Image version: %s", os.Getenv("RELEASE"))
}
