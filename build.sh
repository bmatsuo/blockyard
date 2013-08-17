#!/bin/bash

export GOPATH=${PWD}:$GOPATH

go get ./src/cmd/blockd
go build -o bin/blockd ./src/cmd/blockd
