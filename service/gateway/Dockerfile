FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/gateway ./service/gateway

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/gateway /app/gateway
COPY service/gateway/etc /app/etc

EXPOSE 8086
ENTRYPOINT ["/app/gateway", "-f", "/app/etc/gateway.yaml"]
