#!/bin/sh
set -e

echo "📦 Verifying migration files exist..."
ls -la ./db/migrations

echo "🚀 Running database migrations..."
# Try running standard migration
./chess-server migrate || ./chess-server migrate --dir ./db/migrations || ./chess-server migrate -path ./db/migrations

echo "🌐 Starting Chess Server..."
exec ./chess-server