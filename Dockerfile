FROM scratch
WORKDIR /
COPY ./dist /.

ENTRYPOINT ["/capostman"]
