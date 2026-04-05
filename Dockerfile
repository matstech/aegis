# Copyright 2026-present matstech
# SPDX-License-Identifier: GPL-3.0-only

FROM registry.access.redhat.com/ubi9/go-toolset:1.25.7 AS builder

WORKDIR /opt/app-root/src

ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=go1.26.1+auto

COPY server ./server
COPY security ./security
COPY configuration ./configuration

COPY go.mod ./go.mod
COPY go.sum ./go.sum
COPY main.go ./main.go
COPY version.go ./version.go

RUN go build \
    -ldflags "-X main.BUILDDATE=$(date +%Y-%m-%dT%H:%M:%S%z)" \
    -v -o /tmp/aegis ./main.go

FROM scratch
WORKDIR /app

COPY --from=builder /etc/pki /etc/pki
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /tmp/aegis /app/aegis

USER 1001

ENTRYPOINT ["/app/aegis"]
