#!/bin/bash

export GOPATH=${PWD}:$GOPATH

go build -o bin/blockd ./src/cmd/blockd
