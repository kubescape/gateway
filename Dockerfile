# FROM scratch
# WORKDIR /
# COPY ./dist /.

# ENTRYPOINT ["./canotificationserver"]

FROM alpine:3.11

RUN apk update
RUN apk add ca-certificates 

WORKDIR /
COPY ./dist /.

ENTRYPOINT ["./canotificationserver"]
