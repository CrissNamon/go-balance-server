# syntax=docker/dockerfile:1
FROM golang:latest

WORKDIR /app

COPY ./ /app

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT go mod tidy && go test -v test/*.go && CompileDaemon -build="go build" -command=./balance-server -directory=. -exclude-dir=.git -exclude=".#*" -polling=true
