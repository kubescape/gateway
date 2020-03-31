package main

import (
	"canotificationserver/notificationserver"
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	fmt.Printf("main")
	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	ns.SetupNotificationServer()
}
