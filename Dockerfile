FROM scratch
WORKDIR /
ADD websock_notify /main
EXPOSE 8001
EXPOSE 8002
ENTRYPOINT ["/main"]
