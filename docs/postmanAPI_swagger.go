/*
Package docs documents the HTTP API endpoints.
The documentation is then used to generate an OpenAPI spec.
*/
package docs

import ns "github.com/armosec/cluster-notifier-api-go/notificationserver"

// Notification IDs
//
// The IDs of the sent notifications
//
// Example: [1, 2, 3]
type notificationIDs []int

/*
A request to send a notification has been successfully received.

swagger:response postSendNotificationOk
*/
type postSendNotificationOk struct {
	// In: body
	Body notificationIDs
}

/*
A request to send a notification is malformed

swagger:response postSendNotificationBadRequest
*/
type postSendNotificationBadRequest struct {
	// In: body
	Body string
}

/*
swagger:parameters postSendNotification
*/
type postSendNotificationParams struct {
	// In: body
	Body ns.Notification
}

/*
swagger:route POST /v1/sendnotification postSendNotification
Send a notification to the listeners

Responses:
  200: postSendNotificationOk
  400: postSendNotificationBadRequest
*/
