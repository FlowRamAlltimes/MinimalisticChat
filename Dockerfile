FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod init chat
RUN go build -o chat-server maybe.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/chat-server .
EXPOSE 8080
CMD ["./chat-server"]
