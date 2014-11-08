GOPATH := ${PWD}
SOURCES := $(wildcard src/*.go src/launch/*.go)
PREFIX ?= /usr/local

all: sbin/launch_socket_server

sbin/launch_socket_server: $(SOURCES)
	GOPATH=${GOPATH} go build -o sbin/launch_socket_server src/launch_socket_server.go

install: sbin/launch_socket_server
	mkdir -p ${PREFIX}/sbin ${PREFIX}/libexec/launch_socket_server
	cp -p sbin/launch_socket_server ${PREFIX}/sbin
	cp -p libexec/launch_socket_server/login_wrapper ${PREFIX}/libexec/launch_socket_server

clean:
	GOPATH=${GOPATH} go clean
	rm -f sbin/launch_socket_server
