FROM golang:latest

ADD . /go/src/github.com/staffbase/observatory-exporter
WORKDIR /go/src/github.com/staffbase/observatory-exporter

RUN go install -v ./...

ENTRYPOINT  [ "/go/bin/observatory-exporter" ]
EXPOSE      9229
