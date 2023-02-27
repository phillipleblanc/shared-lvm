FROM golang:1.20 AS builder

WORKDIR /go/src/github.com/phillipleblanc/sharedlvm

COPY . .

RUN go build ./cmd/sharedlvm

FROM alpine:3.14

COPY --from=builder /go/src/github.com/phillipleblanc/sharedlvm/sharedlvm /usr/local/bin/sharedlvm

RUN mkdir /csi

ENTRYPOINT ["/usr/local/bin/sharedlvm"]