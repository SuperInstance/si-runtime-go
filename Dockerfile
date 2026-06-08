# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN go build -v ./...
RUN go test -v ./...

# Runtime stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/si-runtime-go .
CMD ["./si-runtime-go"]
