#!/bin/sh

prepare() {
	mkdir -p built-binaries/
	rm -f built-binaries/*
}

buildhstats() {
	local os="$1"
	local arch="$2"
	local suffix oname

	[ "$os" = "windows" ] && suffix=".exe"
	oname="built-binaries/hstats-$os-$arch$suffix"
	GOOS="$os" GOARCH="$arch" \
		go build -ldflags="-s -w" -o "$oname" hstats.go
}

gocrossloop() {
	local goos="windows linux darwin"
	local goarch="386 amd64 arm"

	for os in $goos ; do
		for arch in $goarch ; do
			[ "$arch" = "arm" -a "$os" = "windows" ] && continue
			[ "$arch" = "arm" -a "$os" = "darwin" ] && continue

			"$@" "$os" "$arch"
		done
	done
}

hstats() {
	prepare
	gocrossloop buildhstats
}

# crossing hstats needs crossing go itself first.
crossgo() {
	gocrossloop ./make.bash --no-clean
}

# cross, then
"${1:-hstats}"
