FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o chat-server server.go  


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/chat-server .
EXPOSE 8080
CMD ["./chat-server"]
