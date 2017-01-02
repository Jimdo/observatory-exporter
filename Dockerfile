FROM golang:alpine

ADD . /go/src/github.com/Jimdo/observatory-exporter
WORKDIR /go/src/github.com/Jimdo/observatory-exporter

RUN go install -v ./...

ENTRYPOINT  [ "/go/bin/observatory-exporter" ]
EXPOSE      9229
