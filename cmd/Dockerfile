FROM golang:1.20-alpine as builder
RUN apk --no-cache add git gcc libc-dev linux-headers

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x && mkdir -p ./dist

COPY cmd ./cmd
COPY pkg ./pkg
COPY config ./config
COPY *.go ./

ARG APP_VERSION="0.0.0-dev.0"
ARG APP_NAME
RUN CGO_ENABLED=0 go build -v \
    -ldflags "-X github.com/chronicleprotocol/oracle-suite.Version=${APP_VERSION}" \
    -o ./dist/chronicle-app ./cmd/${APP_NAME}

# ----------------------------------------------------------------------------------------
FROM ghcr.io/chronicleprotocol/watchdog:0.2.0 as watchdog
# ----------------------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian11

COPY --from=watchdog /usr/local/bin/watchdog /usr/local/bin/
COPY --from=builder /app/dist/chronicle-app /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/watchdog","/usr/local/bin/chronicle-app"]
CMD ["run"]

# libp2p
EXPOSE 8000

# webapi
EXPOSE 8080
