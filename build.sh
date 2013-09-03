#! /bin/bash

go build fiddler.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o fiddler.linux fiddler.go
