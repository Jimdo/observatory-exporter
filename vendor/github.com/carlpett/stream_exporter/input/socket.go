package input

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"

	"github.com/prometheus/common/log"
)

type SocketInput struct {
	family     string
	listenAddr string
}

var (
	socketFamily     = flag.String("input.socket.family", "tcp", "Socket family (tcp/udp/etc)")
	socketListenAddr = flag.String("input.socket.listenaddr", "", "Listening address of socket")
)

func init() {
	registerInput("socket", newSocketInput)
}

func newSocketInput() (StreamInput, error) {
	if *socketFamily == "" {
		return nil, errors.New("-input.socket.family not set")
	} else if *socketFamily != "tcp" && *socketFamily != "udp" && *socketFamily != "domain" {
		return nil, errors.New(fmt.Sprintf("%q is not a valid value for -input.socket.family", *socketFamily))
	}

	if *socketListenAddr == "" {
		return nil, errors.New("-input.socket.listenaddr not set")
	}

	return SocketInput{
		family:     *socketFamily,
		listenAddr: *socketListenAddr,
	}, nil
}

func (socket SocketInput) StartStream(ch chan<- string) {
	l, err := net.Listen(socket.family, socket.listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			// TODO: Timeout + metrics for failed reads
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				ch <- scanner.Text()
			}
			c.Close()
		}(conn)
	}
}
