FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/comment/rpc
RUN go build -trimpath -ldflags="-s -w" -o /out/comment-rpc .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/comment-rpc /app/comment-rpc
COPY --from=builder /src/service/comment/rpc/etc /app/etc

EXPOSE 8083

CMD ["/app/comment-rpc", "-f", "/app/etc/comment-rpc.yaml"]
