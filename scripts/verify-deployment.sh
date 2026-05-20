#!/bin/sh
set -e

cp .env.example .env 2>/dev/null || true

docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d

echo 'Waiting for backend...'
for i in $(seq 1 30); do
  curl -sf http://localhost:8080/health && break
  sleep 2
done

echo 'Backend healthy. Running smoke test...'
curl -sf http://localhost/

docker compose -f docker-compose.prod.yml down
echo 'Deployment verified OK'
