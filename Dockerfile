FROM golang:1.16.3 as builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY ./main.go  ./

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build \
    -o /go/bin/datadog-monitor-prometheus-exporter \
    -ldflags '-s -w'

FROM alpine:3.13.5 as runner

COPY --from=builder /go/bin/datadog-monitor-prometheus-exporter /app/datadog-monitor-prometheus-exporter

RUN adduser -D -S -H exporter

USER exporter

ENTRYPOINT ["/app/datadog-monitor-prometheus-exporter"]
