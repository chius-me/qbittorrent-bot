FROM golang:1.24-alpine AS builder

ARG GOPROXY
ENV GOPROXY=${GOPROXY:-https://goproxy.cn,direct}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o qbittorrent-bot .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/qbittorrent-bot .

ENV STATE_FILE=/data/notify_state.json
VOLUME ["/data"]

ENTRYPOINT ["./qbittorrent-bot"]
