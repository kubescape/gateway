package main

import (
	"capostman/notificationserver"
	"flag"
)

func main() {
	flag.Parse()

	ns := notificationserver.NewNotificationServer()
	ns.SetupNotificationServer()
}
