#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

echo "🚀 Running database migrations..."
./chess-server migrate

echo "🌐 Starting Chess Server..."
exec ./chess-server