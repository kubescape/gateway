#!/usr/bin/env bash
set -x

if [ -z "$IMAGE_NAME"]; then
    export IMAGE_NAME="notification-server"
fi

if [ -z "$IMAGE_TAG"]; then
    export IMAGE_TAG="test"
fi

imglist=$(docker images $IMAGE_NAME:$IMAGE_TAG| grep $IMAGE_NAME)

if [ ! -z $imglist ]; then
    echo "removing image $IMAGE_NAME:$IMAGE_TAG"
    docker image rm -f $IMAGE_NAME:$IMAGE_TAG
fi

set -e
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-server ../
docker build --no-cache -t $IMAGE_NAME:$IMAGE_TAG .

rm notification-server
