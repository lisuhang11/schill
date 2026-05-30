FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/relation/rpc
RUN go build -trimpath -ldflags="-s -w" -o /out/relation-rpc .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/relation-rpc /app/relation-rpc
COPY --from=builder /src/service/relation/rpc/etc /app/etc

EXPOSE 8081

CMD ["/app/relation-rpc", "-f", "/app/etc/relation-rpc.yaml"]
