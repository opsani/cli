FROM golang:1.14-alpine

WORKDIR /go/src/app
COPY main.go .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]
