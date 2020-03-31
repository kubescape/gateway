FROM alpine
# FROM scratch
WORKDIR /
# COPY ./dist /.
COPY ./canotificationserver /

ENTRYPOINT ["./canotificationserver"]
