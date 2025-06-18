# syntax=docker/dockerfile:1.4

# ---- Build Stage ----
FROM golang:1.24.1-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /workspace
COPY ./ /workspace

RUN go mod download
RUN go mod tidy
RUN go build -ldflags "-s -w" -o goapp

# ---- Minimal Runtime Stage ----
FROM alpine:3.21

# 安装基础证书支持（让程序能访问 HTTPS）
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /workspace/goapp .

USER 65530
EXPOSE 3000
ENTRYPOINT ["./goapp"]
