FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/interaction/rpc
RUN go build -trimpath -ldflags="-s -w" -o /out/interaction-rpc .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/interaction-rpc /app/interaction-rpc
COPY --from=builder /src/service/interaction/rpc/etc /app/etc

EXPOSE 8084

CMD ["/app/interaction-rpc", "-f", "/app/etc/interaction-rpc.yaml"]
