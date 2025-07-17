FROM --platform=linux/amd64 golang:1.22 AS builder

ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

RUN apt-get update && apt-get install -y gcc g++ librdkafka-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/server/main.go


FROM --platform=linux/amd64 debian:bookworm-slim

RUN apt-get update && apt-get install -y librdkafka1 ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8000
CMD ["./main"]
