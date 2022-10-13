FROM golang:1.19-alpine as builder

WORKDIR /go/src

ENV GOOS=linux
ENV GOARCH=amd64

COPY go.mod main.go ./
RUN go mod tidy
RUN go build -o app

FROM scratch
WORKDIR /go/src

COPY --from=builder /go/src/app /go/src/app

EXPOSE 8080

ENTRYPOINT ["/go/src/app"]