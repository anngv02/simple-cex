# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files from root
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./backend/
COPY engine/ ./engine/
COPY api/ ./api/

# Build the application
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/backend main.go db.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/bin/backend .

# Expose port
EXPOSE 8010

# Run the binary
CMD ["./backend"]

