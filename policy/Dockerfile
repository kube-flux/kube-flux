FROM golang:1.15-alpine AS builder
WORKDIR /

RUN apk add --no-cache git
RUN go get github.com/boltdb/bolt/...

COPY . .
RUN go mod init github.com/kube-flux/kube-flux/policy
RUN go build -o /build/zeus ./main/main.go
ENTRYPOINT ["build/zeus"]