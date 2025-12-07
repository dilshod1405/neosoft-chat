#!/bin/sh

echo "📄 Generating Swagger docs..."
swag init -g cmd/server/main.go --dir ./ -o ./docs

echo "🚀 Starting Chat Service..."
/app/chat
