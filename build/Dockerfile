FROM golang:1.12.10-alpine3.10

ENV GOPATH /go

COPY . /go/src/github.com/lxs137/droidvirt-ctrl

RUN go build -o /operator /go/src/github.com/lxs137/droidvirt-ctrl/cmd/manager

FROM alpine:3.10

COPY --from=0 /operator /usr/local/bin/droidvirt-operator
