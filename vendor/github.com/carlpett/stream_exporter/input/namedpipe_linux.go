package input

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"syscall"

	"github.com/prometheus/common/log"
)

func init() {
	registerInput("namedpipe", newNamedPipeInput)
}

var (
	pipePath = flag.String("input.namedpipe.path", "", "Path where pipe should be created")
)

func newNamedPipeInput() (StreamInput, error) {
	if *filePath == "" {
		return nil, errors.New("-input.namedpipe.path not set")
	}

	return NamedPipeInput{
		path: *pipePath,
	}, nil
}

func (input NamedPipeInput) StartStream(ch chan<- string) {
	err := syscall.Mkfifo(input.path, 0666)
	if err != nil {
		log.Fatal(err)
	}

	pipe, err := os.OpenFile(input.path, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(input.path)

	reader := bufio.NewReader(pipe)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		ch <- scanner.Text()
	}
}
