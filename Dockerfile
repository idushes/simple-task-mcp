# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build main application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# Build create-admin utility
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o create-admin ./cmd/create-admin

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/create-admin .
# Copy migration files
COPY --from=builder /app/database/migrations ./database/migrations

CMD ["./main"]
