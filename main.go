package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"notification-server/notificationserver"

	"github.com/armosec/capacketsgo/k8sshared/probes"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "1")
	DisplayBuildTag()
	isReadinessReady := false
	go probes.InitReadinessV1(&isReadinessReady)
	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	isReadinessReady = true
	ns.SetupNotificationServer()
}

// DisplayBuildTag display on startup
func DisplayBuildTag() {
	imageVersion := "unknown build"
	dat, err := ioutil.ReadFile("./build_number.txt")
	if err == nil {
		imageVersion = string(dat)
	}
	glog.Infof("Image version: %s", imageVersion)
}
