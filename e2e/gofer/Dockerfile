FROM golang:1-alpine as builder
RUN apk --no-cache add git
WORKDIR /go/src/gofer
COPY . .
ENV CGO_ENABLED=0 \
    GOSUMDB=off

RUN rm -rf dist \
    && mkdir dist \
    && go mod vendor \
    && go build -o dist/gofer ./cmd/gofer

FROM golang:1-alpine
RUN apk --no-cache add ca-certificates git
COPY --from=builder /go/src/gofer/dist/ /usr/local/bin/

ENV CGO_ENABLED=0 \
    GOSUMDB=off

ENV SMOCKER_HOST="http://smocker"
ENV SMOCKER="${SMOCKER_HOST}:8080"

COPY . /app
WORKDIR /app
RUN go mod download

CMD ["go", "test", "-v", "-parallel", "1", "-cpu", "1", "./e2e/gofer"]
