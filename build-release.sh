#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o secrets-bridge-linux-amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o secrets-bridge-windows-amd64.exe
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o secrets-bridge-darwin-amd64
