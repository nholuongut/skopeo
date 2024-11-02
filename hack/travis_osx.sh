#!/usr/bin/env bash
set -e

export GOPATH=$(pwd)/_gopath
export PATH=$GOPATH/bin:$PATH

_nholuongut="${GOPATH}/src/github.com/nholuongut"
mkdir -vp ${_nholuongut}
ln -vsf $(pwd) ${_nholuongut}/skopeo

go version
GO111MODULE=off go get -u github.com/cpuguy83/go-md2man golang.org/x/lint/golint

cd ${_nholuongut}/skopeo
make validate-local test-unit-local binary-local
sudo make install
skopeo -v
