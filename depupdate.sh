#!/bin/bash

go get -u ./...

# pin some dependencies
go get -u cuelang.org/go@v0.8.0-rc.1
go get cuelang.org/go/mod/modcache@v0.8.0-rc.1

