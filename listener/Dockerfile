FROM golang:1.17.8-alpine

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go get
RUN go build -o crawler

EXPOSE 8080
ENTRYPOINT ["/build/listener"]