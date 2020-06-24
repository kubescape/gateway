FROM scratch

RUN apk update
RUN apk add ca-certificates 

WORKDIR /
COPY ./notification-server /

ENTRYPOINT ["./notification-server"]
