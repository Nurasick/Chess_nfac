FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./chess-server ./main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /chess-server /chess-server
EXPOSE 8080
ENTRYPOINT ["/chess-server"]

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/chess-server .

# Copy the entrypoint script into the container
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Use ENTRYPOINT to handle the migration flow
ENTRYPOINT ["./entrypoint.sh"]