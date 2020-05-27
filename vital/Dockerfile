FROM golang:1.14-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache git

WORKDIR /go/src/app

# Build modules first for cache speed
COPY go.mod go.sum ./
RUN go mod download

# Build the app
COPY main.go .
RUN go build -o /go/bin/app

# Set our pwd to /app for external mounts
WORKDIR /app

# /go/bin is on the path
CMD ["app"]
