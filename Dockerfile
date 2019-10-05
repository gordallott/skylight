FROM golang:1.12

RUN apt install git

ENV GO111MODULE=on

ENV CONFDIR=/config
VOLUME /config

WORKDIR /go/src/github.com/gordallott/skylight
COPY . .

RUN go mod download
RUN go build

ENTRYPOINT ["go", "run", "."]
