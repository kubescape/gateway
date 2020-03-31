#!/usr/bin/env bash
set -ex

mkdir ./dest
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./dest/capostman .
chmod +x ./dest/capostman
docker build --no-cache -t dreg.eust0.cyberarmorsoft.com:443/capostman:t0 .


rm -rf ./dest

docker push dreg.eust0.cyberarmorsoft.com:443/capostman:t0