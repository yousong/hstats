#!/bin/sh

GOOS="windows linux"
GOARCH="386 amd64 arm"

mkdir -p built-binaries/
rm -f built-binaries/*
for os in $GOOS ; do
	for arch in $GOARCH ; do
		if [ "$arch" = "arm" -a "$os" = "windows" ] ; then
			continue
		fi
		suffix=""
		[ "$os" = "windows" ] && suffix=".exe"
		oname="built-binaries/hstats-$os-$arch$suffix"
		GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w" -o "$oname" hstats.go
	done
done
