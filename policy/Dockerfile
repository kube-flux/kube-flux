FROM golang:1.15-alpine AS builder
WORKDIR /

RUN apk add --no-cache git
RUN go get github.com/boltdb/bolt/...
COPY . .
RUN go mod init github.com/kube-flux/kube-flux/policy
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /build/zeus ./main/main.go

FROM scratch
COPY --from=builder /build/zeus /zeus
ENTRYPOINT ["/zeus"]