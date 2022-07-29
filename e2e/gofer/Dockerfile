FROM golang:1.18-alpine as builder
RUN apk --no-cache add git
WORKDIR /go/src/gofer
COPY . .
ENV CGO_ENABLED=0 \
    GOSUMDB=off

# NOTE splitting each command into its own RUN statement allows better insight
# when a specific command fails
RUN rm -rf dist
RUN mkdir dist
RUN go mod vendor
RUN go build -o dist/gofer ./cmd/gofer
RUN go build -o dist/rpc-splitter ./cmd/rpc-splitter

FROM golang:1-alpine
RUN apk --no-cache add ca-certificates git bash
COPY --from=builder /go/src/gofer/dist/ /usr/local/bin/

ENV CGO_ENABLED=0 \
    GOSUMDB=off

ENV SMOCKER_HOST="http://smocker"
ENV SMOCKER="${SMOCKER_HOST}:8080"

COPY ./e2e/gofer /app
WORKDIR /app
RUN go mod vendor

CMD ["./scripts/test.sh"]
