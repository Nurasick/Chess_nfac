FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files 
COPY . .

# Build the application binary cleanly into /app/chess-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chess-server ./main.go

FROM alpine:3.19
# Install ca-certificates and dos2unix to fix Windows script line-endings
RUN apk --no-cache add ca-certificates dos2unix
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/chess-server .

# Copy and fix the entrypoint script
COPY entrypoint.sh .
RUN dos2unix entrypoint.sh
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]