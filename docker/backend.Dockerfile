FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the source code
COPY backend .
COPY frontend/src/index.html .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/db ./db
COPY --from=builder /app/index.html .

# Expose port
EXPOSE 8080

CMD ["./main"]
