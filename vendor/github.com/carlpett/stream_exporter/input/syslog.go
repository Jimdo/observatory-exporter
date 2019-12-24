package input

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/prometheus/common/log"
	"github.com/valyala/fasttemplate"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type SyslogInput struct {
	listenAddr   string
	listenFamily string
	format       format.Format
	template     *fasttemplate.Template
}

var (
	syslogListenFamily = flag.String("input.syslog.listenfamily", "", "Listening protocol family (tcp/udp/unix)")
	syslogListenAddr   = flag.String("input.syslog.listenaddr", "", "Listening address of syslog server")
	syslogFormatFlag   = flag.String("input.syslog.format", "autodetect", "Format of incoming syslog data (rfc3164/rfc5424/rfc6587/autodetect)")
	syslogLineTemplate = flag.String("input.syslog.linetemplate", "[message][content]", "Template for data to pass to pattern matcher")
)

func init() {
	registerInput("syslog", newSyslogInput)
}

func newSyslogInput() (StreamInput, error) {
	if *syslogListenFamily == "" {
		return nil, errors.New("-input.syslog.listenfamily not set")
	} else if *syslogListenFamily != "tcp" && *syslogListenFamily != "udp" && *syslogListenFamily != "unix" {
		return nil, errors.New(fmt.Sprintf("%q is not a valid value for -input.syslog.listenfamily", *syslogListenFamily))
	}

	if *syslogListenAddr == "" {
		return nil, errors.New("-input.syslog.listenaddr not set")
	}

	var syslogFormat format.Format
	switch *syslogFormatFlag {
	case "":
		syslogFormat = syslog.Automatic
	case "autodetect":
		syslogFormat = syslog.Automatic
	case "rfc3164":
		syslogFormat = syslog.RFC3164
	case "rfc5424":
		syslogFormat = syslog.RFC5424
	case "rfc6587":
		syslogFormat = syslog.RFC6587
	default:
		return nil, errors.New(fmt.Sprintf("%q is not a valid value for -input.syslog.format", *syslogFormatFlag))
	}

	return SyslogInput{
		listenAddr:   *syslogListenAddr,
		listenFamily: *syslogListenFamily,
		format:       syslogFormat,
		template:     fasttemplate.New(*syslogLineTemplate, "[", "]"),
	}, nil
}

func (input SyslogInput) StartStream(ch chan<- string) {
	syslogChannel := make(syslog.LogPartsChannel)
	logHandler := syslog.NewChannelHandler(syslogChannel)

	server := syslog.NewServer()

	var err error
	switch input.listenFamily {
	case "tcp":
		err = server.ListenTCP(input.listenAddr)
	case "udp":
		err = server.ListenUDP(input.listenAddr)
	case "unix":
		err = server.ListenUnixgram(input.listenAddr)
	default:
		log.Fatalf("Unknown listen family %q", input.listenFamily)
	}
	if err != nil {
		log.Fatal(err)
	}

	server.SetHandler(logHandler)
	server.SetFormat(input.format)

	err = server.Boot()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Syslog server started listening at %s", input.listenAddr)

	go messageHandler(input.template, ch, syslogChannel)

	server.Wait()
	log.Info("Syslog server shutting down")
}

func messageHandler(template *fasttemplate.Template, ch chan<- string, lineIn syslog.LogPartsChannel) {
	for parts := range lineIn {
		line := template.ExecuteFuncString(func(w io.Writer, tag string) (int, error) { return tmplTagFunc(w, tag, parts) })
		ch <- line
	}
}

func tmplTagFunc(w io.Writer, tag string, m map[string]interface{}) (int, error) {
	v := m[tag]
	if v == nil {
		return 0, nil
	}
	switch value := v.(type) {
	case []byte:
		return w.Write(value)
	case string:
		return w.Write([]byte(value))
	case int:
		return w.Write([]byte(strconv.FormatInt(int64(value), 10)))
	case time.Time:
		return w.Write([]byte(value.String()))
	case fmt.Stringer:
		return w.Write([]byte(value.String()))
	default:
		return w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}
