package notificationserver

import (
	"capostman/notificationserver"
	"fmt"
)

func main() {

	ns := notificationserver.NewNotificationServer()
	fmt.Printf("NewNotificationServer")
	ns.SetupNotificationServer()
}
