# ---------- Stage 1: Builder ----------
FROM golang:1.22 AS builder

WORKDIR /app

# 預先加載模組依賴（加快 build）
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼
COPY . .

# 編譯 main 程式
RUN go build -o app ./cmd/server

# ---------- Stage 2: Runtime ----------
FROM debian:bookworm-slim

# 安裝 Kafka 所需的 librdkafka 執行時函式庫與憑證
RUN apt-get update && apt-get install -y \
    librdkafka1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 從 builder 拷貝編譯完成的程式
COPY --from=builder /app/app .

# 執行
CMD ["./app"]
    