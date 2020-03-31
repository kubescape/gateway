package main

import (
	"canotificationserver/notificationserver"
	"fmt"
)

func main() {

	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	ns.SetupNotificationServer()
}
