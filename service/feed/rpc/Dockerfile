FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/feed/rpc
RUN go build -trimpath -ldflags="-s -w" -o /out/feed-rpc .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/feed-rpc /app/feed-rpc
COPY --from=builder /src/service/feed/rpc/etc /app/etc

EXPOSE 8087

CMD ["/app/feed-rpc", "-f", "/app/etc/feed-rpc.yaml"]
