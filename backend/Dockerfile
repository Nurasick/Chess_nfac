FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chess-server ./main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates dos2unix
WORKDIR /app

COPY --from=builder /app/chess-server .

# 📦 COPY THE MIGRATIONS FOLDER INTO THE FINAL IMAGE
COPY db/migrations ./db/migrations

COPY entrypoint.sh .
RUN dos2unix entrypoint.sh
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]