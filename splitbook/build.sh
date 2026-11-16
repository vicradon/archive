#!/bin/bash
set -e
mkdir -p bin
GOOS=windows GOARCH=amd64 go build -o bin/splitbook-windows-amd64.exe main.go
GOOS=darwin GOARCH=amd64 go build -o bin/splitbook-darwin-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o bin/splitbook-darwin-arm64 main.go
GOOS=linux GOARCH=amd64 go build -o bin/splitbook-linux-amd64 main.go
GOOS=linux GOARCH=arm64 go build -o bin/splitbook-linux-arm64 main.go
