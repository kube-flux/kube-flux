FROM golang:1.13-alpine AS builder
WORKDIR /build
COPY hello_world.go .
RUN apk add --no-cache git
RUN go get github.com/boltdb/bolt/...
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o hello_world hello_world.go

FROM scratch
COPY --from=builder /build/hello_world /hello_world
ENTRYPOINT ["/hello_world"]
