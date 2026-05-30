FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /src/service/search/api
RUN go build -trimpath -ldflags="-s -w" -o /out/search-api .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/search-api /app/search-api
COPY --from=builder /src/service/search/api/etc /app/etc

EXPOSE 8895

CMD ["/app/search-api", "-f", "/app/etc/search-api.yaml"]
