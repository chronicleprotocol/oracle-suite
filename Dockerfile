FROM golang:1.18-alpine as builder
RUN apk --no-cache add git gcc libc-dev linux-headers

WORKDIR /go/src/oracle-suite

COPY . .

RUN go mod tidy && go mod vendor

ARG CGO_ENABLED=1
RUN go build -o ./dist/ ./cmd/...

#### Final Image ####
FROM alpine:3.16
RUN apk --no-cache add ca-certificates bash

COPY --from=builder /go/src/oracle-suite/dist/* /usr/local/bin/

ENV HOME=/usr/share/oracle-suite/
WORKDIR ${HOME}

EXPOSE 8000
EXPOSE 8082
EXPOSE 9000
EXPOSE 9100

COPY ./config.json .

CMD ["/bin/bash"]
