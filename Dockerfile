# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o llm-gateway ./cmd/gateway

# 运行阶段
FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /app/llm-gateway /usr/local/bin/
COPY configs/config.example.yaml /etc/llm-gateway/config.yaml

EXPOSE 8080

ENTRYPOINT ["llm-gateway"]
CMD ["-c", "/etc/llm-gateway/config.yaml"]
