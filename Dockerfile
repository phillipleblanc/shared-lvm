FROM golang:1.20 AS builder

WORKDIR /go/src/github.com/phillipleblanc/sharedlvm

COPY . .

RUN go build ./cmd/sharedlvm

FROM ubuntu:22.04

RUN apt-get update && apt-get install ca-certificates musl-dev -y
RUN apt-get install lvm2 -y

COPY --from=builder /go/src/github.com/phillipleblanc/sharedlvm/sharedlvm /usr/local/bin/sharedlvm

RUN mkdir /csi

ENTRYPOINT ["/usr/local/bin/sharedlvm"]