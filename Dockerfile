FROM golang:1-alpine as builder
RUN apk --no-cache add git gcc libc-dev linux-headers

WORKDIR /go/src/oracle-suite

COPY . .

RUN go mod tidy && go mod vendor

ARG CGO_ENABLED=1

RUN go build -o ./dist/ ./cmd/...

FROM alpine:3
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/oracle-suite/dist/* /usr/local/bin/
