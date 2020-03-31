package main

import (
	"canotificationserver/notificationserver"
	"flag"
)

func main() {
	flag.Parse()

	ns := notificationserver.NewNotificationServer()
	ns.SetupNotificationServer()
}
