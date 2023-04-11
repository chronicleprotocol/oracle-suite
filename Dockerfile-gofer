FROM golang:1.20-alpine as builder
RUN apk --no-cache add git
WORKDIR /go/src/gofer
COPY . .
RUN    export CGO_ENABLED=0 \
    && mkdir -p dist \
    && go mod vendor \
    && go build -o dist/gofer ./cmd/gofer

FROM alpine:3.16
RUN apk --no-cache add ca-certificates 
WORKDIR /root
COPY --from=builder /go/src/gofer/dist/ /usr/local/bin/
COPY ./config.hcl ./config.hcl
EXPOSE 9200
ENTRYPOINT ["/usr/local/bin/gofer"]
CMD ["-c", "config.hcl", "agent"]
