# ---- 构建阶段：交叉编译 linux 二进制 ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/miaoverse .

# ---- 运行阶段：最小镜像 ----
FROM alpine:3.20

# 时区数据与常用调试工具
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

COPY --from=builder /out/miaoverse /app/miaoverse
COPY --from=builder /src/locales /app/locales

# 容器内配置文件由外部挂载注入，避免把密钥打进镜像
EXPOSE 3000

ENTRYPOINT ["/app/miaoverse"]
