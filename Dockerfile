FROM golang:1.22-alpine AS builder
WORKDIR /build

RUN apk add --no-cache git

COPY go.mod ./
COPY go.su[m] ./

RUN go mod download || go mod tidy

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o feature_orbit_server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/feature_orbit_server ./feature_orbit_server
EXPOSE 8080
ENTRYPOINT ["./feature_orbit_server"]
