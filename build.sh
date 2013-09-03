#! /bin/bash

go build -o bin/fiddler fiddler.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/fiddler.linux fiddler.go
tar -czvf bin/fiddler.tar.gz bin/fiddler
tar -czvf bin/fiddler.linux.tar.gz bin/fiddler.linux