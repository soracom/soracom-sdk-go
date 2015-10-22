#!/bin/bash

go fmt
go vet
go test $*

