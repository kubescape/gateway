package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"notification-server/notificationserver"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "1")
	DisplayBuildTag()

	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
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
