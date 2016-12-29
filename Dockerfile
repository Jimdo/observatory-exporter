
FROM golang:alpine

ADD . /go/src/github.com/Jimdo/observatory-exporter
WORKDIR /go/src/github.com/Jimdo/observatory-exporter

RUN go install -v ./...

COPY observatory-exporter  /bin/observatory-exporter

ENTRYPOINT  [ "/bin/observatory-exporter" ]
EXPOSE      9229