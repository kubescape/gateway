package main

import (
	"fmt"
	"notification-server/notificationserver"
)

func main() {

	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	ns.SetupNotificationServer()
}
