FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/replayor ./cmd/replayor/main.go

WORKDIR /app

FROM golang:1.23-alpine

# Install certs
# RUN apk add --no-cache ca-certificates
# COPY ./oplabs.crt /usr/local/share/ca-certificates/oplabs.crt
# RUN chmod 644 /usr/local/share/ca-certificates/oplabs.crt && update-ca-certificates

COPY --from=builder /app/bin/replayor /app/bin/replayor

CMD ["/app/bin/replayor"]