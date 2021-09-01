FROM golang:1.17 AS builder
COPY . /tmp/src
RUN cd /tmp/src && \
    GOOS=linux GOARCH=amd64 go build -buildmode=plugin -trimpath -o /tmp/tcp-audit-pgsql-sink.so && \
    chmod 400 /tmp/tcp-audit-pgsql-sink.so

FROM scratch
COPY --from=builder /tmp/tcp-audit-pgsql-sink.so /tmp/tcp-audit-pgsql-sink.so
ENTRYPOINT []