FROM golang:1.22-alpine AS builder
WORKDIR /build

# 安装构建工具
RUN apk add --no-cache git

# 单独拷贝 go.mod，go.sum 可能不存在所以用通配符却不会报错
COPY go.mod ./
COPY go.su[m] ./

# 如果 go.sum 不存在则自动生成；已存在则直接用缓存加速
RUN go mod download || go mod tidy

# 拷贝全部源码
COPY . .

# 确保 go.sum 存在（源码拷贝后再跑一次 tidy，保证一致性）
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ums ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/ums        ./ums
COPY --from=builder /build/configs    ./configs
COPY --from=builder /build/migrations ./migrations
EXPOSE 8081
ENTRYPOINT ["./ums"]
