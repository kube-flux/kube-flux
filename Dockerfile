FROM golang:1.14-alpine AS build

WORKDIR /src/
COPY go_helloworld.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/demo

FROM scratch
COPY --from=build /bin/demo /bin/demo
ENTRYPOINT ["/bin/demo"]
