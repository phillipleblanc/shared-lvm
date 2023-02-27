FROM golang:1.20 AS builder

WORKDIR /go/src/github.com/phillipleblanc/sharedlvm

COPY . .

RUN go build ./cmd/sharedlvm

FROM alpine:3.14

RUN apk add --no-cache ca-certificates libc6-compat
RUN apk add --no-cache lvm2 lvm2-extra util-linux device-mapper

COPY --from=builder /go/src/github.com/phillipleblanc/sharedlvm/sharedlvm /usr/local/bin/sharedlvm

RUN mkdir /csi

ENTRYPOINT ["/usr/local/bin/sharedlvm"]