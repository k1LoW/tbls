FROM golang:1-bookworm AS builder

WORKDIR /workdir/
COPY . /workdir/

RUN apt-get update && apt-get install -y sqlite3

RUN update-ca-certificates

RUN make build

FROM debian:bookworm-slim

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workdir/tbls ./usr/bin

ENTRYPOINT ["/entrypoint.sh"]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
