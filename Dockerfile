FROM scratch
COPY ./bin/opsani /
ENTRYPOINT ["/opsani"]
