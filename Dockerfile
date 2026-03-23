# 构建阶段
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/gateway/main.go

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/gateway .
# 如果需要配置文件，取消下面行的注释并确保 configs 目录存在
# COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./gateway"]