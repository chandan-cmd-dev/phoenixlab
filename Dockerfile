FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for go modules
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o phoenixlab ./cmd/main.go

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/phoenixlab .

# Copy configuration and templates
COPY --from=builder /app/conf      ./conf
COPY --from=builder /app/views     ./views
COPY --from=builder /app/static    ./static
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./phoenixlab"]
