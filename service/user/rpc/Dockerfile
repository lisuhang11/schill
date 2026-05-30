FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/user/rpc
RUN go build -trimpath -ldflags="-s -w" -o /out/user-rpc .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/user-rpc /app/user-rpc
COPY --from=builder /src/service/user/rpc/etc /app/etc

EXPOSE 8080

CMD ["/app/user-rpc", "-f", "/app/etc/user-rpc.yaml"]
