#!/bin/bash

# MentalChat Build Script
# Usage: ./build.sh [--build] [--run]

set -e

BUILD=false
RUN=false

# Parse arguments
for arg in "$@"
do
    case $arg in
        --build)
            BUILD=true
            shift
            ;;
        --run)
            RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $arg"
            echo "Usage: ./build.sh [--build] [--run]"
            exit 1
            ;;
    esac
done

# Default to build if no arguments provided
if [ "$BUILD" = false ] && [ "$RUN" = false ]; then
    BUILD=true
    RUN=true
fi

echo "🚀 MentalChat Build Script"
echo "=========================="

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Check if .env.json exists
if [ ! -f ".env.json" ]; then
    echo "⚠️  Warning: .env.json not found. Copying .env.example.json..."
    cp .env.example.json .env.json
    echo "📝 Please configure .env.json with your settings"
fi

# Build backend
if [ "$BUILD" = true ]; then
    echo "🔨 Building Backend..."
    cd Backend
    go mod tidy
    go build -o ../mentalchat-backend ./cmd/main.go
    cd ..
    echo "✅ Backend built successfully"
fi

# Build frontend
if [ "$BUILD" = true ]; then
    echo "🔨 Building Frontend..."
    cd Frontend
    npm install
    npm run build
    cd ..
    echo "✅ Frontend built successfully"
fi

# Run
if [ "$RUN" = true ]; then
    echo "▶️  Running MentalChat..."
    
    # Start backend
    echo "Starting backend server..."
    ./mentalchat-backend &
    BACKEND_PID=$!
    
    # Start frontend
    echo "Starting frontend server..."
    cd Frontend
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    
    echo "✅ MentalChat is running!"
    echo "🌐 Frontend: http://localhost:3000"
    echo "🔧 Backend: http://localhost:8080"
    
    # Trap to kill processes on exit
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null" EXIT
    
    # Wait for processes
    wait
fi

echo "✨ Build script completed"
