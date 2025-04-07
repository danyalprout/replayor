FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/replayor ./cmd/replayor/main.go

WORKDIR /app

FROM golang:1.22

COPY --from=builder /app/bin/replayor /app/bin/replayor

CMD ["/app/bin/replayor"]