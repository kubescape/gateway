# Gateway

The Gateway enables broadcasting a single message to the different microservices running in the cluster.

The gateway works as a tree: there is the root and the different leafs.
The leafs “attach” to the root using a set of attributes, while the root provides an API.

When broadcasting a message, the message must contain the attributes to whom it should be broadcast.
The root will broadcast the message to all the leafs that registered with those attributes.

<img src=".out/design.gif">


## API Documentation

As mentioned before, the Gateway exposes an HTTP API.
You can learn more about the API using one of the provided interactive OpenAPI UIs:
- SwaggerUI, available at `/openapi/v2/swaggerui`
- RapiDoc, available at `/openapi/v2/rapi`
- Redoc, available at `/openapi/v2/docs`


## Supported environment variables
* `CA_NOTIFICATION_SERVER_WS_PORT`: websocket port (default `8001`)
* `CA_NOTIFICATION_SERVER_PORT`: restAPI port (default `8002`)

For more details on environment variables, check out notificationserver/environmentvariables.go.
