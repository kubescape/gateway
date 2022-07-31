FROM golang:1.18-alpine as builder



ENV GO111MODULE=

ENV CGO_ENABLED=0

WORKDIR /work
ADD . .

RUN apk add git

WORKDIR /work
RUN go build -o build/notification

FROM alpine

RUN addgroup -S ks && adduser -S ks -G ks
USER ks
WORKDIR /home/ks/

COPY --from=builder /work/build/notification /usr/bin/notification
ENV RELEASE=$image_version

ENTRYPOINT ["notification"]
