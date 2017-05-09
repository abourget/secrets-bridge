#!/bin/bash


echo ""
echo "Building for Linux"
echo ""
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o secrets-bridge-linux-amd64

echo ""
echo "Building for Windows"
echo ""
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o secrets-bridge-windows-amd64.exe

echo ""
echo "Building for Darwin"
echo ""
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o secrets-bridge-darwin-amd64
