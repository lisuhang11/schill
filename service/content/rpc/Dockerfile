FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/content/rpc
RUN go build -trimpath -ldflags="-s -w" -o /out/content-rpc .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/content-rpc /app/content-rpc
COPY --from=builder /src/service/content/rpc/etc /app/etc

EXPOSE 8082

CMD ["/app/content-rpc", "-f", "/app/etc/content-rpc.yaml"]
