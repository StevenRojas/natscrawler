FROM golang:1.17.8-alpine as builder

WORKDIR /build
RUN mkdir /build/app

COPY ../grpcapi /build/grpcapi
COPY ./crawler/go.mod /build/app
COPY ./crawler/go.sum /build/app
WORKDIR /build/app
RUN go mod download

COPY ./crawler/ /build/app
RUN go build -o crawler cmd/main.go

ENTRYPOINT ["./crawler"]

#ENTRYPOINT ["ping", "google.com"]