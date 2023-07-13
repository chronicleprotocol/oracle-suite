FROM golang:1.20-alpine as builder
RUN apk --no-cache add git gcc libc-dev linux-headers

WORKDIR /app
COPY . .
RUN go mod vendor && mkdir -p dist

ARG APP_VERSION="0.0.0-dev.0"
ARG APP_NAME
RUN CGO_ENABLED=0 go build -v \
    -ldflags "-X github.com/chronicleprotocol/oracle-suite/cmd.Version=${APP_VERSION}" \
    -o ./dist/chronicle-app ./cmd/${APP_NAME}

# ----------------------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/dist/* /usr/local/bin/

COPY ./config*.hcl /usr/share/chronicle-suite/

WORKDIR /usr/share/chronicle-suite/

ENTRYPOINT ["/usr/local/bin/chronicle-app"]
CMD ["run"]

# libp2p
EXPOSE 8000

# webapi
EXPOSE 8080
