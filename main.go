package main

import (
	"fmt"
	"notification-server/notificationserver"
)

func main() {
	// DisplayBuildTag()

	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	ns.SetupNotificationServer()
}

// // DisplayBuildTag display on startup
// func DisplayBuildTag() {
// 	imageVersion := "unknown build"
// 	dat, err := ioutil.ReadFile("./build_number.txt")
// 	if err == nil {
// 		imageVersion = string(dat)
// 	} else {
// 		dat, err = ioutil.ReadFile("./build_date.txt")
// 		if err == nil {
// 			imageVersion = fmt.Sprintf("%s, date: %s", imageVersion, string(dat))
// 		}
// 	}
// 	log.Printf("Image version: %s", imageVersion)
// }
