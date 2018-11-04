FROM golang:1.9 AS builder

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN mkdir -p /go/src/github.com/***
WORKDIR /go/src/github.com/***

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

ADD . /go/src/github.com/benguild/gke-grpc-example
WORKDIR /go/src/github.com/benguild/gke-grpc-example
RUN dep ensure -vendor-only
RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go install github.com/benguild/gke-grpc-example

FROM alpine:latest
COPY --from=0 /go/bin/gke-grpc-example .
CMD ["./gke-grpc-example"]
