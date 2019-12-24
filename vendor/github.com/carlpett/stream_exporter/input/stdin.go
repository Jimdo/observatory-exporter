package input

import (
	"bufio"
	"flag"
	"os"
)

var (
	writeOnEOF = flag.Bool("input.stdin.write-on-eof", false, "If all metrics should be written to stdout after seeing end of file")
	quitOnEOF  = flag.Bool("input.stdin.quit-on-eof", false, "If the exporter should exit after seeing end of file")
)

type StdinInput struct {
}

func init() {
	registerInput("stdin", newStdinInput)
}

func newStdinInput() (StreamInput, error) {
	return StdinInput{}, nil
}

func (input StdinInput) StartStream(ch chan<- string) {
	reader := bufio.NewReader(os.Stdin)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		ch <- scanner.Text()
	}

	if *writeOnEOF {
		writeMetrics(os.Stdout)
	}
	if *quitOnEOF {
		close(ch)
	}
}
