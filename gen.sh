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
	local osarch
	local os arch

	go tool dist list | while read osarch; do
		echo "working on $osarch"
		os="${osarch%/*}"
		arch="${osarch#*/}"
		"$@" "$os" "$arch"
	done
}

hstats() {
	prepare
	gocrossloop buildhstats
}

# cross, then
"${1:-hstats}"
