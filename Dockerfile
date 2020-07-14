FROM alpine:3.11

RUN apk update
RUN apk add ca-certificates 

WORKDIR /
COPY ./dist /.

COPY ./build_number.txt /
RUN echo $(date -u) > /build_date.txt

ENTRYPOINT ["./notification-server"]
