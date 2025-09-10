# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
# Create default .env.example if it doesn't exist in the build context
RUN echo '# Database\nDATABASE_URL=postgres://user:password@postgres:5432/task_manager?sslmode=disable\n\n# MCP Server\nMCP_SERVER_PORT=8080\n\n# JWT\nJWT_SECRET=your-secret-key-here\n\n# Logging\nLOG_LEVEL=info' > .env.example
CMD ["./main"]
