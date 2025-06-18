# syntax=docker/dockerfile:1.4

# ---- Build Stage ----
FROM golang:1.24.1-alpine AS builder

# 使用国内代理提升下载速度
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=1

# 安装 C 编译工具链和 libwebp
RUN apk add --no-cache build-base libwebp-dev

WORKDIR /workspace
COPY ./ /workspace

RUN go mod download
RUN go mod tidy
RUN go build -ldflags "-s -w" -o goapp

# ---- Minimal Runtime Stage ----
FROM alpine:3.21

# 安装基础证书支持和 libwebp 运行时库（某些平台需要）
RUN apk add --no-cache ca-certificates libwebp

WORKDIR /app
COPY --from=builder /workspace/goapp .

USER 65530
EXPOSE 3000
ENTRYPOINT ["./goapp"]
