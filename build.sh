#!/usr/bin/env bash
set -ex
docker image rm dreg.eust0.cyberarmorsoft.com:443/canotificationserver:t0
# mkdir -p ./dist
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o canotificationserver .
docker build --no-cache -t dreg.eust0.cyberarmorsoft.com:443/canotificationserver:t0 .

rm canotificationserver
# docker push dreg.eust0.cyberarmorsoft.com:443/notificationserver:t3