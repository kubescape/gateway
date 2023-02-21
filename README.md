# Gateway
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkubescape%2Fgateway.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkubescape%2Fgateway?ref=badge_shield)


The Gateway enables broadcasting a single message to the different microservices running in the cluster.

The gateway works as a tree: there is the root and the different leafs.
The leafs “attach” to the root using a set of attributes, while the root provides an API.

When broadcasting a message, the message must contain the attributes to whom it should be broadcast.
The root will broadcast the message to all the leafs that registered with those attributes.

<img src=".out/design/design.gif">

## Building gateway
To build the gateway run: `go build .`  

## Configuration
Load config file using the `CONFIG` environment variable   

`export CONFIG=path/to/clusterData.json`  

<details><summary>example/clusterData.json</summary>

```json5 
{
   "gatewayWebsocketURL": "127.0.0.1:8001",
   "gatewayRestURL": "127.0.0.1:8002",
   "kubevulnURL": "127.0.0.1:8081",
   "kubescapeURL": "127.0.0.1:8080",
   "eventReceiverRestURL": "https://report.armo.cloud",
   "eventReceiverWebsocketURL": "wss://report.armo.cloud",
   "rootGatewayURL": "wss://ens.euprod1.cyberarmorsoft.com/v1/waitfornotification",
   "accountID": "*********************",
   "clusterName": "******" 
  } 
``` 
</details>

## API Documentation

As mentioned before, the Gateway exposes an HTTP API.
You can learn more about the API using one of the provided interactive OpenAPI UIs:
- SwaggerUI, available at `/openapi/v2/swaggerui`
- RapiDoc, available at `/openapi/v2/rapi`
- Redoc, available at `/openapi/v2/docs`


## Supported environment variables
* `WEBSOCKET_PORT`: websocket port (default `8001`)
* `HTTP_PORT`: restAPI port (default `8002`)

For more details on environment variables, check out `pkg/environmentvariables.go`.

## VS code configuration samples

You can use the sample file below to setup your VS code environment for building and debugging purposes.

<details><summary>.vscode/launch.json</summary>

```json5
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program":  "${workspaceRoot}",
                 "env": {
                     "NAMESPACE": "armo-system",
                     "CONFIG": "${workspaceRoot}/.vscode/clusterData.json",
            },
            "args": [
                "-alsologtostderr", "-v=4", "2>&1"
            ]
        }
    ]
}
```
</details>


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkubescape%2Fgateway.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkubescape%2Fgateway?ref=badge_large)