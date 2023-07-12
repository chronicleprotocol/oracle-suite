FROM golang:1.20-alpine as builder
RUN apk --no-cache add git gcc libc-dev linux-headers

WORKDIR /app
COPY . .
RUN go mod vendor && mkdir -p dist

ARG APP_VERSION="0.0.0-dev.0"
ARG APP_NAME
RUN CGO_ENABLED=0 go build -v \
    -ldflags "-X github.com/chronicleprotocol/oracle-suite/cmd.Version=${APP_VERSION}" \
    -o ./dist/chronicle-oracle-toolkit ./cmd/${APP_NAME}

# ----------------------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian11

WORKDIR /root
COPY --from=builder /app/dist/* /usr/local/bin/

ARG APP_NAME
COPY ./config-*.hcl ./cmd/${APP_NAME}/config.hcl /usr/share/oracle-suite/

ENTRYPOINT ["/usr/local/bin/chronicle-oracle-toolkit"]
CMD ["-c", "/usr/share/oracle-suite/config.hcl", "run"]

# libp2p
EXPOSE 8000

# webapi
EXPOSE 8080
