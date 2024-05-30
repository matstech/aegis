FROM golang:1.22-alpine AS builder


RUN apk update && apk add --no-cache bash build-base gcc ca-certificates \
    tzdata g++ libc-dev pkgconf && update-ca-certificates

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    usr

COPY server ./server
COPY security ./security
COPY configuration ./configuration

COPY go.mod ./go.mod
COPY go.sum ./go.sum
COPY main.go ./main.go
COPY version.go ./version.go
# COPY config.json ./config.json


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl \
    -ldflags "-extldflags -static -X main.BUILDDATE=`date +%Y-%m-%dT%T%z`" \
    -v -o /go/bin/go-token-guard main.go

FROM scratch
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/go-token-guard /app/go-token-guard

USER usr:usr

ENTRYPOINT ["/app/go-token-guard"]