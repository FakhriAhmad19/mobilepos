FROM golang:1.25-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /build
COPY . .
RUN go mod tidy
RUN go build --ldflags "-s -w -extldflags -static" -o main .

FROM alpine:latest

# tzdata + certs for correct time handling and outbound TLS.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /www

COPY --from=builder /build/main /www/
COPY --from=builder /build/.env /www/.env
COPY --from=builder /build/public/ /www/public/
COPY --from=builder /build/resources/ /www/resources/

# Writable dirs Goravel expects at runtime (logs, framework cache).
RUN mkdir -p /www/storage/logs /www/storage/framework

EXPOSE 3000

ENTRYPOINT ["/www/main"]
