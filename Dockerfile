FROM scratch
ADD websock_notify /capostman
EXPOSE 8001
EXPOSE 8002
CMD ["./capostman"]