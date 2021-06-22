FROM golang:1.16

WORKDIR /go/src/vito
COPY vito/vito.go ./vito/
COPY pkg/ ./pkg/
COPY go.mod go.sum ./

RUN go get -d -v ./...
RUN go install -v ./...

RUN mkdir -p /go/src/vito/
RUN touch /go/src/vito/config.toml

CMD ["vito"]