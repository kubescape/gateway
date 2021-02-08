#!/usr/bin/env bash
set -ex

# export ITAG=latest
export WTAG=test #broadcom-v4

# dep ensure
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-server .
chmod +x notification-server

docker build --no-cache -f Dockerfile.test -t quay.io/armosec/notification-server-ubi:$WTAG .
rm -rf notification-server
# docker push quay.io/armosec/notification-server-ubi:$WTAG

echo "update k8s-notification-server"

kubectl -n cyberarmor-system patch  deployment ca-notification-server -p '{"spec": {"template": {"spec": { "containers": [{"name": "ca-notification-server", "imagePullPolicy": "Never"}]}}}}' || true
kubectl -n cyberarmor-system set image deployment/ca-notification-server ca-notification-server=quay.io/armosec/notification-server-ubi:$WTAG || true
kubectl -n cyberarmor-system delete pod $(kubectl get pod -n cyberarmor-system | grep ca-notification-server |  awk '{print $1}')
kubectl -n cyberarmor-system logs -f $(kubectl get pod -n cyberarmor-system | grep ca-notification-server |  awk '{print $1}')
