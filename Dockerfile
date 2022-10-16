FROM golang:1.19-alpine

WORKDIR /go/src

ENV GOOS=linux
ENV GOARCH=amd64

COPY go.mod main.go ./
RUN go mod tidy
RUN go build -o app

ENTRYPOINT ["/go/src/app"]