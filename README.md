# ARMO In-Cluster notifier

Using the ARMO in-cluster notifier enables broadcasting a single message to the different ARMO microservices running in the cluster.

The notifier works as a tree- there is the root and the different leafs.
The leafs 'register' to the root using a set of attributes, while the root has an open API.

When broadcasting a message, the message must contain the attributes to whom it should be broadcast, the root will broadcast the message to all the leafs that registered with those attributes.

This architecture enables multiple layers of leafs for a ...

<img src=".out/design.gif">


### Supported environment variables
* `CA_NOTIFICATION_SERVER_WS_PORT`: websocket port (default `8001`)
* `CA_NOTIFICATION_SERVER_PORT`: restAPI port (default `8002`)