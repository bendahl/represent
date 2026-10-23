#!/bin/bash

go build -ldflags "-X main.version=$(git describe --abbrev=0 --always) -X main.commit=$(git rev-parse HEAD)"
