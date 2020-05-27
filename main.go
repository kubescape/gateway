package main

import (
	"fmt"
	"notificationserver/notificationserver"
)

func main() {

	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	ns.SetupNotificationServer()
}
