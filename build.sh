#! /bin/bash

go build -o bin/fiddler fiddler.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/fiddler.linux fiddler.go
cd bin
tar -czvf fiddler.tar.gz fiddler
tar -czvf fiddler.linux.tar.gz fiddler.linux
cd ..