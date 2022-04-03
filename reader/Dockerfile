FROM golang:1.17.8-alpine

WORKDIR /build
RUN mkdir /build/app

COPY ../grpcapi /build/grpcapi
COPY ./reader/go.mod /build/app
COPY ./reader/go.sum /build/app
WORKDIR /build/app
RUN go mod download

COPY ./reader/ /build/app
RUN go build -o reader cmd/main.go

CMD tail -f /dev/null