# syntax=docker/dockerfile:1

# ============================================================
# 构建阶段：编译出静态二进制
# docs/ 目录（docs.go / swagger.json / swagger.yaml）已提交到仓库，
# 且被 main.go 以 _ "go-service/docs" 导入，因此无需安装 swag、无需 swag init。
# ============================================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# git 供 go mod download 拉取部分依赖使用
RUN apk add --no-cache git

# 先拷贝依赖清单并下载，利用 Docker 层缓存（依赖不变时跳过重新下载）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝全部源码并编译。
# CGO_ENABLED=0：go-sql-driver/mysql 为纯 Go 实现，可静态编译，便于在精简 alpine 运行。
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o server ./cmd/server/...

# ============================================================
# 运行阶段：仅带二进制的精简镜像
# ============================================================
FROM alpine:latest

WORKDIR /app

# ca-certificates：访问微信/又拍云等 HTTPS 接口需要根证书
# tzdata + 软链：容器内时区设为 Asia/Shanghai，日志时间与业务一致
RUN apk add --no-cache ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /app/server .

# 应用默认监听 3000（PORT 环境变量可覆盖，见 pkg/config/config.go）
EXPOSE 3000

CMD ["./server"]
