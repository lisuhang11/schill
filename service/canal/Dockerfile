FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/canal
RUN go build -trimpath -ldflags="-s -w" -o /out/canal .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/canal /app/canal
COPY --from=builder /src/service/canal/etc /app/etc

EXPOSE 8088

CMD ["/app/canal", "-f", "/app/etc/canal.yaml"]
