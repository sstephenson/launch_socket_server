package main

import (
	"fmt"
	"io"
	"launch"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

var programName string
var programArgs []string
var daemonSocketName string
var destinationAddress string
var destinationType string

func main() {
	if len(os.Args) < 2 {
		die("usage: %s <program> [<arg> ...]", os.Args[0])
	}

	programName = os.Args[1]
	programArgs = os.Args[2:]

	daemonSocketName = os.Getenv("LAUNCH_DAEMON_SOCKET_NAME")
	if daemonSocketName == "" {
		daemonSocketName = "Socket"
		os.Setenv("LAUNCH_DAEMON_SOCKET_NAME", daemonSocketName)
	}

	destinationAddress = os.Getenv("LAUNCH_PROGRAM_TCP_ADDRESS")
	if destinationAddress != "" {
		destinationType = "tcp"

	} else {
		destinationAddress = os.Getenv("LAUNCH_PROGRAM_SOCKET_PATH")
		destinationType = "unix"

		if destinationAddress == "" {
			if programName == "-" {
				die("launch_socket_server: please set LAUNCH_PROGRAM_TCP_ADDRESS or LAUNCH_PROGRAM_SOCKET_PATH")
			}

			destinationAddress = generateSocketPath()
			os.Setenv("LAUNCH_PROGRAM_SOCKET_PATH", destinationAddress)
		}
	}

	start()
}

func start() {
	listeners, err := launch.SocketListeners(daemonSocketName)
	if err != nil || len(listeners) == 0 {
		die("launch_socket_server: error activating launch socket: %s", err)
	}

	if programName != "-" {
		go run()
	}

	for _, listener := range listeners {
		go serve(listener)
	}

	select {}
}

func run() {
	cmd := exec.Command(programName, programArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		die("launch_socket_server: program `%s' exited: %s", programName, err)
	}

	os.Exit(0)
}

func serve(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			die("launch_socket_server: error accepting connection: %s", err)
		}

		go handle(conn)
	}
}

func handle(in net.Conn) {
	out, err := net.Dial(destinationType, destinationAddress)
	if err != nil {
		warn("launch_socket_server: error connecting to program: %s", err)
		in.Close()
		return
	}

	go proxy(in, out)
	go proxy(out, in)
}

func proxy(in net.Conn, out net.Conn) {
	io.Copy(in, out)
	in.Close()
	out.Close()
}

func generateSocketPath() string {
	name := "launch_socket_server.sock-" + strconv.Itoa(os.Getpid())
	return filepath.Join(os.TempDir(), name)
}

func warn(specifier string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, specifier+"\n", values...)
}

func die(specifier string, values ...interface{}) {
	warn(specifier, values...)
	os.Exit(1)
}
