FROM golang:1.17.8-alpine

WORKDIR /build
RUN mkdir /build/app

COPY ../grpcapi /build/grpcapi
COPY ./listener/go.mod /build/app
COPY ./listener/go.sum /build/app
WORKDIR /build/app
RUN go mod download

COPY ./listener/ /build/app
RUN go build -o listener cmd/main.go

EXPOSE 7777

ENTRYPOINT ["./listener"]