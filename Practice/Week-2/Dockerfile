FROM golang:1.14-alpine
WORKDIR /src/
COPY main.go go.* /src/
RUN go build -o main main.go
EXPOSE 8080
CMD ["/src/main", "/v1/myapp"]

